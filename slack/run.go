package slack

import (
	"log"
	"os"

	"backlink/notion"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func Run(appToken, botToken string, session *notion.Session) {
	log.Println("running slack bot")

	api := slack.New(
		botToken,
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	go func() {
		for evt := range client.Events {
			log.Println("e...")
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to slack with socket mode...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %+v\n", evt)

					continue
				}
				log.Printf("Event received: %+v\n", eventsAPIEvent)

				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						_, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
						log.Printf("bot mentioned")
						if err != nil {
							log.Printf("failed posting message: %v", err)
						}
					case *slackevents.MessageEvent:
						log.Printf("msg sent")
						HandleMsgs(ev, client, api, session)
					}
				default:
					client.Debugf("unsupported Events API event received")
				}
			}
		}

	}()

	err := client.Run()
	if err != nil {
		log.Println(err)
	}
}
