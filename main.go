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
)

var MsgCh = make(chan bool, 1)
var DscCh = make(chan bool, 1)
var BtCh = make(chan bool, 1)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = "sounds"

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
	}

	msg := <-MsgCh
	if msg {
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		log.Println("PlayAudioFile:", "kaboom")
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

	} else {
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

	TargetUser = m.Author

	gs, err := s.Guild(GuildID)
	if err != nil {
		log.Fatal(err)
	}

	switch m.Content {
	case "!kaboom":
		if searchVoiceStates(gs.VoiceStates, m.Author.ID) {
			MsgCh <- true
		} else {
			MsgCh <- false
		}
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

func createMentions(users []*discordgo.User) string {
	var mentions []string
	for _, u := range users {
		mentions = append(mentions, u.Mention())
	}
	return strings.Join(mentions, " ")
}
