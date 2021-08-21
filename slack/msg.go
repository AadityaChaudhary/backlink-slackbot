package slack

import (
	"log"
	"regexp"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func HandleMsgs(ev *slackevents.MessageEvent, client *socketmode.Client) {
	backlinks := getBacklinks(ev.Text)
	log.Println(backlinks)
}

func getBacklinks(msg string) []string {
	r, _ := regexp.Compile(`\[\[([^]]+)\]\]`)
	b := r.FindAllString(msg, -1)
	for i, v := range b {
		b[i] = v[2 : len(v)-2]
	}
	return b
}
