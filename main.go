package main

import (
	"fmt"
	"log"
	"os"

	"backlink/slack"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("hello")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	log.Println("app", slackAppToken)
	log.Println("bot", slackBotToken)

	slack.Run(slackAppToken, slackBotToken)
}
