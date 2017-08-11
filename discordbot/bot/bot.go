package bot

import (
	"fmt"

	"strings"

	"../config"
	"github.com/bwmarrin/discordgo"
)

// ID for the bot
var BotID string
var goBot *discordgo.Session

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID

	goBot.AddHandler(messageHandler)
	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("The bot is running!")

	<-make(chan struct{})
	return
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if string.SplitAfter(m.Content, " ") == 
		if m.Content == "!addproject" {
			_, _ = s.ChannelMessageSend(m.ChannelID, "no")
		}
	}
}
