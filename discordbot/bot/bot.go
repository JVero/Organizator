package bot

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	"strings"

	"../botdatabase"

	"../config"
	"github.com/bwmarrin/discordgo"
)

// BotID is the ID for the bot
var BotID string
var goBot *discordgo.Session

// Project is a struct containing the database format
type Project struct {
	Contributors []string `bson:"Contributors"`
	Name         string   `bson:"Name"`
	Creator      string   `bson:"Creator"`
}

// Start starts the bot (duh)
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Session establishes connection to database
	databaseSession := botdatabase.Start()
	defer databaseSession.Close()
	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID

	goBot.AddHandler(defaultHandler)
	goBot.AddHandler(dbFetchContributor(databaseSession))
	goBot.AddHandler(dbFetchCreator(databaseSession))

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("The bot is running!")

	<-make(chan struct{})
	return
}

func defaultHandler(d *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!findcontributors") {
			fmt.Println("This happened")
		}
		if strings.SplitAfter(m.Content, " ")[0] == "!addproject" {
			_, _ = d.ChannelMessageSend(m.ChannelID, "ok")
		}

	}
}

func dbFetchContributor(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!findcontributors") {
			name := strings.Split(m.Content, " ")[1]
			contributors := botdatabase.GetContributorsByName(db, name)

			for index, contributor := range contributors {
				contributors[index] = "<@" + string(contributor) + ">"
			}
			message := "The contributors to " + name + " are:\n " + strings.Join(contributors, "\n")

			_, _ = d.ChannelMessageSend(m.ChannelID, message)

		}
	}
}

func dbFetchCreator(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!findcreator") {
			name := strings.Split(m.Content, " ")[1]
			creator := botdatabase.GetCreatorByName(db, name)

			message := "The creator of " + name + " is <@" + creator + ">"

			_, _ = d.ChannelMessageSend(m.ChannelID, message)

		}
	}
}
