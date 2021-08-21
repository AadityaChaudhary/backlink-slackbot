package slack

import (
	"log"
	"regexp"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func HandleMsgs(ev *slackevents.MessageEvent, client *socketmode.Client, api *slack.Client) {
	backlinks := getBacklinks(ev.Text)

	log.Println(ev.Message)
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

func GetTeamName(api *slack.Client) (string, error) {
	resp, err := api.AuthTest()
	if err != nil {
		return "", err
	}
	return resp.Team, nil
}
