package main

import (
	"fmt"
	"log"
	"os"

	"backlink/notion"
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
	notionToken := os.Getenv("NOTION_SECRET")
	log.Println("app", slackAppToken)
	log.Println("bot", slackBotToken)
	log.Println("notion", notionToken)

	client := notion.NewClient(notionToken)
	session, err := notion.NewSession(client, []string{os.Getenv("B_PARENT")})
	if err != nil {
		log.Println(err)
	}

	slack.Run(slackAppToken, slackBotToken, &session)
	log.Println("notion")
}
