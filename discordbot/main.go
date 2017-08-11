package main

import (
	"fmt"

	"./bot"
	"./config"
)

const token string = ""

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bot.Start()

	<-make(chan struct{})
	return
}
