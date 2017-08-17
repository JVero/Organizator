package bot

import (
	"fmt"
	"strconv"
	"strings"

	"../botdatabase"
	"../config"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
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
	goBot.AddHandler(helpHandler)

	goBot.AddHandler(dbFetchContributors(databaseSession))
	goBot.AddHandler(dbFetchCreator(databaseSession))
	goBot.AddHandler(dbAddContributor(databaseSession))
	goBot.AddHandler(dbAddProject(databaseSession))
	goBot.AddHandler(dbFetchProjects(databaseSession))
	goBot.AddHandler(dbAddPermissions(databaseSession))
	goBot.AddHandler(dbRemoveProject(databaseSession))

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
		if strings.HasPrefix(m.Content, "!findcontributors ") {
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
		if strings.HasPrefix(m.Content, "!getcontributors ") {
			name := strings.Replace(m.Content, "!getcontributors ", "", 1)
			contributors := botdatabase.GetContributorsByName(db, name)
			if len(contributors) == 0 {
				message := "There are no contributors for that project yet, or more likely that project doesn't exist.  Use !getprojects to see a list of active projects"
				_, _ = d.ChannelMessageSend(m.ChannelID, message)
				return

			}
			for index, contributor := range contributors {
				contributors[index] = "<@" + string(contributor) + ">"
			}
			message := "The contributors to ``" + name + "`` are:\n " + strings.Join(contributors, "\n")

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
		if strings.HasPrefix(m.Content, "!getcreator ") {

			projectName := strings.Replace(m.Content, "!getcreator ", "", 1)

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
		if strings.HasPrefix(m.Content, "!addme ") {
			if botdatabase.HasPermission(db, m.Author.ID) {
				newContributor := m.Author.ID
				projectName := strings.Replace(m.Content, "!addme ", "", 1)

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
				return
			}
			message := "User is not allowed to make this request"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}

	}
}

func dbAddProject(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotID {
			return
		}
		if strings.HasPrefix(m.Content, "!addproject ") {
			if botdatabase.HasPermission(db, m.Author.ID) {
				fmt.Println(m.MentionRoles)

				newProject := strings.Replace(m.Content, "!addproject ", "", 1)
				fmt.Println(newProject)
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
				return
			}
			message := "User is not allowed to make this request"
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
		message := "The possible commands for this bot include: ```\n" +
			"_________GETS_____________\n\n" +
			"!getcontributors <project> \n" +
			"!getcreator <project> \n" +
			"!getprojects \n\n" +
			"_________SETS______________\n\n" +
			"!addme <project> \n" +
			"!addproject <project> \n```"
		d.ChannelMessageSend(m.ChannelID, message)
	}
}

func dbAddPermissions(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.HasPrefix(m.Content, "!permit <@") && strings.HasSuffix(m.Content, ">") {
			// Checks if the user sending the command can give permission.  People with permission can give permission
			if botdatabase.HasPermission(db, m.Author.ID) {
				_, err := strconv.Atoi(strings.Split(strings.SplitAfter(m.Content, " <@")[1], ">")[0])
				if err != nil {
					message := "Please don't try to cheat the bot.  He's a good bot"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
				newUser := strings.Split(strings.SplitAfter(m.Content, " <@")[1], ">")[0]

				// Edits the database to add the newUser to the list of allowed
				status := botdatabase.AddPermissions(db, newUser)
				if status == "Already Exists" {
					message := "<@" + newUser + "> already has permission!"
					_, _ = d.ChannelMessageSend(m.ChannelID, message)
					return
				}
				fmt.Println(newUser)
				message := "User <@" + newUser + "> successfully added!"
				_, _ = d.ChannelMessageSend(m.ChannelID, message)
				return
			}
			message := "You do not have appropriate permissions to add a user!"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}
	}
}

func dbRemoveProject(db *mgo.Session) func(d *discordgo.Session, m *discordgo.MessageCreate) {
	return func(d *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.HasPrefix(m.Content, "!remove ") {
			if botdatabase.HasPermission(db, m.Author.ID) {
				projectName := strings.Replace(m.Content, "!remove ", "", 1)
				botdatabase.RemoveProjectFromDatabase(db, projectName)
				message := "The project has been removed!  Confirm by using !getprojects"
				_, _ = d.ChannelMessageSend(m.ChannelID, message)
				return
			}
			message := "You do not have the permissions to do this"
			_, _ = d.ChannelMessageSend(m.ChannelID, message)
		}
	}
}
