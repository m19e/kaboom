package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func init() {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "dev")
	}

	err := godotenv.Load(fmt.Sprintf(".env%s", os.Getenv("GO_ENV")))
	if err != nil {
		log.Fatal(err)
	}
}

var (
	Token      string
	GuildID    string
	TChannelID string
	VChannelID string
	Folder     string
	err        error

	TargetUser *discordgo.User
	Loop       bool
	BGList     []File
)

type File struct {
	Cmd  string
	Name string
}

var MsgCh = make(chan string, 1)
var BgmCh = make(chan string, 10)
var DscCh = make(chan bool, 1)
var BtCh = make(chan bool, 1)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = "sounds"
	Loop = true
	BGList = []File{
		{Cmd: "okini", Name: "お気に入りの場所"}, {Cmd: "kagayaki", Name: "輝きはじめる私達の日常"}, {Cmd: "mattari", Name: "まったりいこうよ"}, {Cmd: "dousitano", Name: "どうしたの？"}, {Cmd: "hotto", Name: "ホッと一息安心感"}, {Cmd: "sawayaka", Name: "爽やかな朝"}, {Cmd: "town", Name: "TOWN(SCmix)"}, {Cmd: "yuudoki", Name: "夕時ノスタルジック"},
	}

	var Sess *discordgo.Session

	// Infinite loop
connectLoop:
	for {
		// Connect to Discord
		Sess, err = discordgo.New("Bot " + Token)
		if err != nil {
			log.Fatal(err)
		}

		// Open Websocket
		err = Sess.Open()
		if err != nil {
			log.Fatal(err)
		}

		// Run process
		if err = run(Sess); err != nil {
			log.Fatal(err)
		}

		select {
		// TODO: Receive session close signal
		case <-DscCh:
			log.Printf("disconnect and break loop.\n")
			_, err = Sess.ChannelMessageSend(TChannelID, "See you.")
			if err != nil {
				log.Fatal(err)
			}
			break connectLoop
		default:
			log.Printf("close session.\n")
			Sess.Close()
		}
	}

	Sess.Close()
}

func run(s *discordgo.Session) error {
	// Register the messageCreate func as a callback
	// NOTE: dont SET handler in LOOP!!!!!
	s.AddHandler(messageCreate)

	// Send boot message only once
	if len(BtCh) == 0 {
		_, err = s.ChannelMessageSend(TChannelID, "!kaboom is Ready.")
		if err != nil {
			return err
		}
		BtCh <- true
		MsgCh <- "cmd"
	}

	msg := <-MsgCh
	switch msg {
	case "cmd":
		plain := "!kaboom\nVCを爆破します\n!karan\n麦茶の氷が鳴ります\n!bg (bgname)\n!bgに続けて曲名を指定することでBGMを流します\n!bglist\n!bgで指定できるBGMの一覧を表示します\n!loop\n!bgでループ再生するかどうかを切り替えます\n!cmd\nコマンド一覧を表示します"

		embed := &discordgo.MessageEmbed{
			Color:  0xff4500,
			Fields: []*discordgo.MessageEmbedField{},
		}
		var field struct {
			Name  string
			Value string
		}

		for i, nv := range strings.Split(plain, "\n") {
			if i%2 == 0 {
				field.Name = nv
			} else {
				field.Value = nv
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:  field.Name,
					Value: field.Value,
				})
			}
		}

		_, err = s.ChannelMessageSendEmbed(TChannelID, embed)
		if err != nil {
			return err
		}

	case "kaboom":
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		log.Println("PlayAudioFile:", msg)
		s.UpdateStatus(0, "detonation...")

		gs, err := s.Guild(GuildID)
		if err != nil {
			return err
		}

		var data discordgo.GuildChannelCreateData
		vc, err := s.State.Channel(VChannelID)

		data.Name = "爆破予定地"
		data.Type = 2
		data.ParentID = vc.ParentID

		c, err := s.GuildChannelCreateComplex(GuildID, data)
		if err != nil {
			return err
		}

		defer func() {
			log.Println("leave VC")
			dgv.Disconnect()
			log.Printf("close voice connection\n")
			dgv.Close()

			_, err = s.ChannelDelete(c.ID)
			if err != nil {
				return
			}
		}()

		target, err := s.ChannelMessageSend(TChannelID, "Please follow me.")
		if err != nil {
			return err
		}

		time.Sleep(2 * time.Second)

		s.ChannelVoiceJoin(GuildID, c.ID, false, true)

		var convicts []*discordgo.User

		for _, vs := range gs.VoiceStates {
			if vs.ChannelID == VChannelID {
				member, err := s.GuildMember(GuildID, vs.UserID)
				if err != nil {
					return err
				}
				convicts = append(convicts, member.User)
			}
		}

		time.Sleep(3 * time.Second)

		for _, cnv := range convicts {
			s.GuildMemberMove(GuildID, cnv.ID, c.ID)
			time.Sleep(250 * time.Millisecond)
		}

		count, err := s.ChannelMessageEdit(TChannelID, target.ID, "Start a countdown.")
		if err != nil {
			return err
		}

		time.Sleep(3 * time.Second)

		for i := 5; i > 0; i-- {
			time.Sleep(1 * time.Second)
			count, err = s.ChannelMessageEdit(TChannelID, count.ID, fmt.Sprintf("%d", i))
			if err != nil {
				return err
			}
		}

		time.Sleep(1 * time.Second)

		count, _ = s.ChannelMessageEdit(TChannelID, count.ID, "KABOOM!")
		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, "kaboom.mp4"), make(chan bool))

		_, err = s.ChannelMessageSend(TChannelID, fmt.Sprintf("%s See you.", createMentions(convicts)))
		if err != nil {
			log.Fatal(err)
		}
	case "karan":
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		defer func() {
			log.Println("leave VC")
			dgv.Disconnect()
			log.Printf("close voice connection\n")
			dgv.Close()
		}()

		log.Println("PlayAudioFile:", msg)
		s.UpdateStatus(0, "melting...")

		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, fmt.Sprintf("%s.mp4", msg)), make(chan bool))
		time.Sleep(2 * time.Second)

	case "cinema":
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		defer func() {
			log.Println("leave VC")
			dgv.Disconnect()
			log.Printf("close voice connection\n")
			dgv.Close()
		}()

		log.Println("PlayAudioFile:", msg)
		s.UpdateStatus(0, "(♪beep)")

		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, fmt.Sprintf("%s.mp4", msg)), make(chan bool))
		time.Sleep(2 * time.Second)

	case "bg":
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		defer func() {
			log.Println("leave VC")
			dgv.Disconnect()
			log.Printf("close voice connection\n")
			dgv.Close()
		}()

		bgm := <-BgmCh

		if !exists(fmt.Sprintf("%s/%s", Folder, bgm)) {
			_, err = s.ChannelMessageSend(TChannelID, "Not found.")
			if err != nil {
				return err
			}
			return nil
		}

		log.Println("PlayAudioFile:", bgm)
		s.UpdateStatus(0, fmt.Sprintf("♪%s", bgm))

		for {
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, bgm), make(chan bool))
			select {
			case bgm = <-BgmCh:

			default:

			}
			if !Loop {
				break
			}
		}
	case "bglist":
		embed := &discordgo.MessageEmbed{
			Color:  0xff4500,
			Fields: []*discordgo.MessageEmbedField{},
		}

		for _, bg := range BGList {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("!bg %s", bg.Cmd),
				Value:  fmt.Sprintf("♪「%s」", bg.Name),
				Inline: false,
			})
		}

		_, err = s.ChannelMessageSendEmbed(TChannelID, embed)
		if err != nil {
			return err
		}

	case "loop":
		Loop = !Loop
		loopSt := func(loop bool) string {
			if loop {
				return "Set loop."
			} else {
				return "Set unloop."

			}
		}(Loop)

		_, err = s.ChannelMessageSend(TChannelID, loopSt)
		if err != nil {
			return err
		}

	case "reject":
		_, err = s.ChannelMessageSend(TChannelID, fmt.Sprintf("%s Please order after join VC.", TargetUser.Mention()))
		if err != nil {
			return err
		}
	}

	// Wait here until Ctrl-C or other term signal is received.
	// fmt.Println("Bot is now running. Press Ctrl-C to exit.")
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	// <-sc

	return nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	gs, err := s.Guild(GuildID)
	if err != nil {
		log.Fatal(err)
	}

	if searchVoiceStates(gs.VoiceStates, m.Author.ID) {
		TargetUser = m.Author

		if strings.HasPrefix(m.Content, "!bg ") {
			MsgCh <- "bg"
			BgmCh <- parseBG(m.Content)
			return
		}

		MsgCh <- m.Content[1:]
	} else {
		MsgCh <- "reject"
	}
}

func searchVoiceStates(vss []*discordgo.VoiceState, id string) bool {
	for _, vs := range vss {
		if vs.UserID == id {
			return true
		}
	}

	return false
}

func parseBG(s string) string {
	bg := s[4:]
	bg = strings.TrimSpace(bg)

	return fmt.Sprintf("%s.mp4", bg)
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func createMentions(users []*discordgo.User) string {
	var mentions []string
	for _, u := range users {
		mentions = append(mentions, u.Mention())
	}
	return strings.Join(mentions, " ")
}
