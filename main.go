package main

import (
	"fmt"
	"log"
	"os"

	"backlink/slack"

	"github.com/joho/godotenv"
)

func main() {
  // test stuff from db branch
  
// 	if err := InitDB(false); err != nil {
// 		fmt.Println(err)
// 		panic(err)
// 	}

// 	//DropAllTables()

// 	//if err := AddWorkspace("nag"); err != nil {
// 	//	fmt.Println(err)
// 	//	panic(err)
// 	//}

// 	//err := AddBacklinkToWorkspace("nag", Backlink{LinkName: "bushan", NotionID: "nagabushan"})
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//	panic(err)
// 	//}

// 	workspace := GetWorkspaceInfo("nag")
// 	fmt.Println(workspace)

// 	defer func() {
// 		if err := DeinitDB(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

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
