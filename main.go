package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = "sounds"

	// Connect to Discord
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal(err)
	}

	// Open Websocket
	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}
connectLoop:
	for {
		if err = run(discord); err != nil {
			log.Fatal(err)
		}

		select {
		case <-DscCh:
			log.Printf("disconnect and break loop.\n")
			break connectLoop
		default:

		}
	}

	log.Printf("close session.\n")
	discord.Close()
}

func run(s *discordgo.Session) error {
	// Register the messageCreate func as a callback
	s.AddHandler(messageCreate)

	_, err = s.ChannelMessageSend(TChannelID, "!kaboom is Ready.")
	if err != nil {
		log.Println("Error sending message: ", err)
	}

	msg := <-MsgCh
	if msg {
		// Connect to voice channel.
		// NOTE: Setting mute to false, deaf to true.
		dgv, err := s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
		if err != nil {
			return err
		}

		fmt.Println("PlayAudioFile:", "hotelmoonside")
		s.UpdateStatus(0, "hotelmoonside")

		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, "hotelmoonside.mp3"), make(chan bool))

		// Close connections
		dgv.Close()
	} else {
		_, err = s.ChannelMessageSend(TChannelID, "See you.")
		if err != nil {
			log.Println("Error sending message: ", err)
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

	switch m.Content {
	case "!kaboom":
		MsgCh <- true
	case "!disconnect":
		MsgCh <- false
		DscCh <- true
	}
}
