package bot

import (
	"fmt"

	"gopkg.in/mgo.v2"

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
	goBot.AddHandler(dbFetchContributors(databaseSession))
	goBot.AddHandler(dbFetchCreator(databaseSession))
	goBot.AddHandler(dbAddContributor(databaseSession))
	goBot.AddHandler(dbAddProject(databaseSession))
	goBot.AddHandler(dbFetchProjects(databaseSession))
	goBot.AddHandler(helpHandler)
	goBot.AddHandler(dbAddPermissions(databaseSession))

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("The bot is running!")

	<-make(chan struct{})
	return
}

// defaultHandler is just a simple template on how handers should work
func defaultHandler(d *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!findcontributors") {
			fmt.Println("This happened")
		}

	}
}

// Fetches the contributors of the project from the database, then formats it for Discord and sends a message
func dbFetchContributors(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!getcontributors") {
			name := strings.Split(m.Content, " ")[1]
			contributors := botdatabase.GetContributorsByName(db, name)
			if len(contributors) == 0 {
				message := "There are no contributors for that project yet :("
				_, _ = d.ChannelMessageSend(m.ChannelID, message)

			}
			for index, contributor := range contributors {
				contributors[index] = "<@" + string(contributor) + ">"
			}
			message := "The contributors to " + name + " are:\n " + strings.Join(contributors, "\n")

			_, _ = d.ChannelMessageSend(m.ChannelID, message)

		}
	}
}

// Fetches the creator of the project from the database, then formats it for Discord and sends a message
func dbFetchCreator(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!getcreator") {

			projectName := strings.Split(m.Content, " ")[1]

			allProjects := botdatabase.FetchAllProjectsFromDatabase(db)
			for _, value := range allProjects {
				if value.Name == projectName {
					message := "The creator of " + value.Name + " is <@" + value.Creator + ">"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
			}
			message := "There is no project with the name " + projectName + ".\n" +
				"To find a list of all projects, use the command !getprojects"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}
	}
}

func dbAddContributor(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!addme") {
			newContributor := m.Author.ID
			projectName := strings.SplitAfter(m.Content, " ")[1]

			previousContributors := botdatabase.GetContributorsByName(db, projectName)
			for _, contributors := range previousContributors {
				if contributors == newContributor {
					message := "That user has already been added to the project"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
			}
			newContributors := append(previousContributors, newContributor)

			allProjects := botdatabase.FetchAllProjectsFromDatabase(db)
			for _, project := range allProjects {
				if project.Name == projectName {
					botdatabase.SetContributorsByName(db, projectName, newContributors)
					message := "User <@" + m.Author.ID + "> successfully added!"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
			}
			message := "That project doesn't exist, use the !getprojects command to see a list of available projects"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)

		}

	}
}

func dbAddProject(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!addproject") {
			newProject := strings.SplitAfter(m.Content, " ")[1]
			projects := botdatabase.FetchAllProjectsFromDatabase(db)
			for _, project := range projects {
				if project.Name == newProject {
					message := "A project already exists with this name, please try another name.  For a list of existing projects, use the command !listprojects"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
			}
			botdatabase.AddProjectToDatabase(db, m.Author.ID, newProject)

			message := "Project successfully added!"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}
	}
}

func dbFetchProjects(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!getprojects") {
			projects := botdatabase.FetchAllProjectsFromDatabase(db)
			message := "```\n"
			for _, project := range projects {
				fmt.Println(project.Name)
				message += project.Name + "\n"
			}
			message += "```"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}
	}
}

func helpHandler(d *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Message.Content == "!help" {
		message := "The possible commands for this bot include ```\n" +
			"!getcontributors <project> \n" +
			"!getcreator <project> \n" +
			"!addme <project> \n" +
			"!addproject <project> \n" +
			"!getprojects \n```"
		d.ChannelMessageSend(m.ChannelID, message)
	}
}

// Not functional yet, TBA feature
/*func dbAddPermissions(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.HasPrefix(m.Content, "!permit") {
			if botdatabase.HasPermission(db, m.Author.ID) == true {
				newUser := strings.SplitAfter(m.Content, " ")[1]
				fmt.Println(newUser)
			}
		}
	}
}*/
