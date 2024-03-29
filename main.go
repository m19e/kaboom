package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var Sounds = map[string]string{}

func init() {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "local")
	}

	err := godotenv.Load(fmt.Sprintf(".env.%s", os.Getenv("GO_ENV")))
	if err != nil {
		log.Println(err)
	}
}

var (
	Token      string
	GuildID    string
	TChannelID string
	VChannelID string
	Folder     string
	err        error

	dgv        *discordgo.VoiceConnection
	jobs       chan string
	playing    bool
	detonation bool
)

var rep = regexp.MustCompile(`<:([^:]+):\d+>`)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = os.Getenv("FOLDER")

	files, _ := ioutil.ReadDir(Folder)

	for _, f := range files {
		filename := f.Name()
		Sounds[strings.Split(filename, ".")[0]] = filename
	}

	jobs = make(chan string, 10)

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal(err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(messageCreate)

	dg.UpdateGameStatus(1, "!join to invite")

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Content == "" || detonation {
		return
	}

	log.Printf("[%s]\n", m.Content)

	gs, err := s.State.Guild(GuildID)

	if !checkCommand(m.Content) {
		if dgv == nil || !dgv.Ready {
			return
		}

		filename, ok := Sounds[checkStamp(m.Content)]
		if !ok {
			return
		}

		jobs <- filename

		if playing {
			return
		}

		playing = true
		for {
			j, ok := <-jobs
			if !ok {
				playing = false
				jobs = make(chan string, 10)
				break
			}
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, j), make(chan bool))
		}

		return
	}

	if m.Content == "!join" {
		if !searchVoiceStates(gs.VoiceStates, m.Author.ID) {
			s.ChannelMessageSend(TChannelID, fmt.Sprintf("%s Please order after join VC.", m.Author.Mention()))
			return
		}

		dgv, err = s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			log.Fatal(err)
		}
		s.UpdateGameStatus(1, "!leave to disconnect")
		return
	}

	if dgv == nil || !dgv.Ready {
		return
	}

	switch m.Content {
	case "!leave":
		if !searchVoiceStates(gs.VoiceStates, m.Author.ID) {
			s.ChannelMessageSend(TChannelID, fmt.Sprintf("%s Please order after join VC.", m.Author.Mention()))
			return
		}
		err = dgv.Disconnect()
		if err != nil {
			log.Fatal(err)
		}
		s.UpdateGameStatus(1, "!join to invite")

	case "!jihou":
		for _, num := range getCountsRing() {
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, fmt.Sprintf("Bell_use%d.ogg", num)), make(chan bool))
		}

	case "!kaboom":
		if !searchVoiceStates(gs.VoiceStates, m.Author.ID) {
			s.ChannelMessageSend(TChannelID, fmt.Sprintf("%s Please order after join VC.", m.Author.Mention()))
			return
		}

		detonation = true
		s.UpdateGameStatus(1, "")

		var data discordgo.GuildChannelCreateData
		vc, err := s.State.Channel(VChannelID)

		data.Name = "爆破予定地"
		data.Type = 2
		data.ParentID = vc.ParentID

		c, err := s.GuildChannelCreateComplex(GuildID, data)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(3 * time.Second)

		dgv, err = s.ChannelVoiceJoin(GuildID, c.ID, false, true)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(3 * time.Second)

		var convicts []*discordgo.User

		for _, vs := range gs.VoiceStates {
			if vs.ChannelID == VChannelID {
				member, _ := s.GuildMember(GuildID, vs.UserID)
				convicts = append(convicts, member.User)
				s.GuildMemberMove(GuildID, member.User.ID, &c.ID)
				time.Sleep(250 * time.Millisecond)
			}
		}

		time.Sleep(3 * time.Second)

		defer func() {
			err = dgv.Disconnect()
			if err != nil {
				log.Fatal(err)
			}

			_, err = s.ChannelDelete(c.ID)
			if err != nil {
				return
			}

			detonation = false
			s.UpdateGameStatus(1, "!join to invite")
		}()

		// var soundNames = []string{"askr_hgcsorry", "hgc_oko", "kit_pya", "kaboom", "hnn_yaha"}
		var soundNames = []string{"kaboom"}
		for _, sn := range soundNames {
			if sn == "kaboom" {
				gifs := []string{
					"https://media.giphy.com/media/146BUR1IHbM6zu/giphy.gif",
					"https://media.giphy.com/media/HhTXt43pk1I1W/giphy.gif",
					"https://media.giphy.com/media/rkkMc8ahub04w/giphy.gif",
					"https://media.giphy.com/media/3oKIPwoeGErMmaI43S/giphy.gif",
				}
				s.ChannelMessageSend(TChannelID, fmt.Sprintf("See you, %s\n%s", createMentions(convicts), choiceRandomOne(gifs)))
			}
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, Sounds[sn]), make(chan bool))
			time.Sleep(1250 * time.Millisecond)
		}
	}
}

func checkCommand(m string) bool {
	switch m {
	case "!join", "!leave", "!kaboom":
		return true
	}
	return false
}

func checkStamp(m string) string {
	match := rep.FindStringSubmatch(m)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func choiceRandomOne(slice []string) string {
	rand.Seed(time.Now().UnixNano())
	return slice[rand.Intn(len(slice))]
}

func createMentions(users []*discordgo.User) string {
	var mentions []string
	for _, u := range users {
		mentions = append(mentions, u.Mention())
	}
	return strings.Join(mentions, " ")
}

func getCountsRing() []int {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	hour := time.Now().In(jst).Hour()

	if hour == 0 {
		hour = 12
	}
	if hour > 12 {
		hour = hour - 12
	}

	slice := make([]int, 12)
	for i := range make([]int, hour) {
		slice = append(slice, (i%2)+1)
	}

	return slice
}

func searchVoiceStates(vss []*discordgo.VoiceState, id string) bool {
	for _, vs := range vss {
		if vs.UserID == id {
			return true
		}
	}

	return false
}
