package botdatabase

import (
	"fmt"

	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Project is the format for each project
type Project struct {
	Contributors []string `bson:"Contributors"`
	Name         string   `bson:"Name"`
	Creator      string   `bson:"Creator"`
}

// Start the database on port 27017, use StartSpecific to choose your own port location
func Start() *mgo.Session {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Println(err.Error())
	}
	return session
}

// StartSpecific starts the database on a specific port, use Start to use the default (27017)
func StartSpecific(portnumber int) *mgo.Session {
	session, err := mgo.Dial("localhost:" + strconv.Itoa(portnumber))
	if err != nil {
		fmt.Println(err.Error())
	}
	return session
}

// AddProjectToDatabase takes the creators name and the name of the project and adds it to the database
func AddProjectToDatabase(sess *mgo.Session, name string, projectname string) {
	collection := sess.DB("DiscordBot").C("ProjectInfo")

	newProject := Project{
		Contributors: []string{},
		Name:         projectname,
		Creator:      name,
	}

	collection.Insert(newProject)
}

// fetchResponseFromDatabase is a generic function that returns the whole response for each function below to parse
func fetchResponseFromDatabase(sess *mgo.Session, name string) Project {
	collection := sess.DB("DiscordBot").C("ProjectInfo")

	var result Project
	err := collection.Find(bson.M{"Name": name}).One(&result)

	if err != nil {
		fmt.Println("Invalid, check the request.  The possible commands are:```",
			"\n!addproject <projectname>m",
			"\n!addme <projectname>,",
			"\n!getcreator <projectname>,",
			"\n!getcontributors <projectname>```")
	}
	return result
}

// FetchAllProjectsFromDatabase fetches...
func FetchAllProjectsFromDatabase(sess *mgo.Session) []Project {
	collection := sess.DB("DiscordBot").C("ProjectInfo")

	var results []Project

	collection.Find(bson.M{}).All(&results)

	return results
}

// GetCreatorByName returns the creator of the project as a string
func GetCreatorByName(sess *mgo.Session, name string) string {
	result := fetchResponseFromDatabase(sess, name)
	return result.Creator
}

// GetContributorsByName returns an array of all the contributors ID's as strings
func GetContributorsByName(sess *mgo.Session, name string) []string {
	result := fetchResponseFromDatabase(sess, name)
	return result.Contributors
}

// SetContributorsByName updates the list of contributors
func SetContributorsByName(sess *mgo.Session, name string, newContributors []string) {
	collection := sess.DB("DiscordBot").C("ProjectInfo")
	update := bson.M{"$set": bson.M{"Contributors": newContributors}}
	collection.Update(bson.M{"Name": name}, update)

}
