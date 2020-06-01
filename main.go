package main

import (
	"fmt"
	"log"
	"os"
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

		if err = run(Sess); err != nil {
			log.Fatal(err)
		}

		select {
		case <-DscCh:
			log.Printf("disconnect and break loop.\n")
			_, err = Sess.ChannelMessageSend(TChannelID, "See you.")
			if err != nil {
				log.Fatal(err)
			}
			break connectLoop
		default:
			log.Println("leave VC only")
			_, err = Sess.ChannelMessageSend(TChannelID, "See you.")
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("close session.\n")
			Sess.Close()
		}
	}

	// Wait here until Ctrl-C or other term signal is received.
	// fmt.Println("Bot is now running. Press Ctrl-C to exit.")
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	// <-sc

	Sess.Close()
}

func run(s *discordgo.Session) error {
	// Register the messageCreate func as a callback
	// TODO: dont SET handler in LOOP!!!!!
	s.AddHandler(messageCreate)

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

		_, err = s.ChannelMessageSend(TChannelID, "Count down.")
		if err != nil {
			return err
		}

		time.Sleep(3 * time.Second)

		count, err := s.ChannelMessageSend(TChannelID, "5")
		if err != nil {
			return err
		}

		for i := 4; i > 0; i-- {
			time.Sleep(1 * time.Second)
			count, err = s.ChannelMessageEdit(TChannelID, count.ID, fmt.Sprintf("%d", i))
		}

		time.Sleep(1 * time.Second)
		count, err = s.ChannelMessageEdit(TChannelID, count.ID, "KABOOM!")

		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, "kaboom.mp4"), make(chan bool))

		// TODO: GuildMemberRemove

		log.Printf("close voice connection\n")
		dgv.Close()
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

	switch m.Content {
	case "!join":

	case "!kaboom":
		MsgCh <- true
		// DscCh <- true

		// case "!bye":
		// 	MsgCh <- false
		// 	DscCh <- true
	}
}
