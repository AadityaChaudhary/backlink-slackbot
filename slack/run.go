package slack

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func Run(appToken, botToken string) {
	log.Println("running slack bot")

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	go func() {

	}()

	client.Run()
}
