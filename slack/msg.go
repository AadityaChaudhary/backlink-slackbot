package slack

import (
	"backlink/notion"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func HandleMsgs(ev *slackevents.MessageEvent, client *socketmode.Client, api *slack.Client, session *notion.Session) {
	var user string
	var txt string
	var userID string
	var time string
	backlinks := getBacklinks(ev.Text)
	if len(backlinks) == 0 {
		log.Println("no backlinks found")
		return
	}
	txt = ev.Text
	userID = ev.User
	time = ev.TimeStamp

	if ev.ThreadTimeStamp != "" {
		log.Println("thread")
		params := &slack.GetConversationRepliesParameters{
			Timestamp: ev.ThreadTimeStamp,
			ChannelID: ev.Channel,
		}
		msgs, _, _, err := api.GetConversationReplies(params)
		if err != nil {
			log.Println("err", err)
			return
		}
		txt = msgs[0].Text
		userID = msgs[0].User
		time = msgs[0].Timestamp
	}

	u, err := api.GetUserInfoContext(context.Background(), userID)
	if err != nil {
		log.Println(err)
		return
	}
	user = u.Profile.RealName
	timeS, err := convertTime(time)
	log.Println(backlinks)
	log.Println(txt)
	log.Println(user)
	log.Println(time)
	log.Println(timeS.UTC())

	header := fmt.Sprint(user, timeS.UTC())

	_, err = createNewBacklinkPage(session, backlinks[0], header, txt)
	if err != nil {
		log.Println("err", err)
		return
	}

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

func convertTime(ut string) (time.Time, error) {
	uts := strings.Split(ut, ".")
	s, err := strconv.Atoi(uts[0])
	ns, err := strconv.Atoi(uts[1])
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(s), int64(ns)), nil
}

func createNewBacklinkPage(session *notion.Session, title, header, para string) (string, error) {

	blocks := []notion.Block{
		notion.Block{
			Object: "block",
			Type:   "heading_3",
			Heading3: &notion.Text{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: header,
						},
					},
				},
			},
		},
		notion.Block{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &notion.TextTree{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: para,
						},
					},
				},
			},
		},
	}
	p, err := session.Pages[0].AppendPageWithBlocks(title, blocks)
	return p.Id, err
}

func addContent(session *notion.Session, pageID, header, para string) error {
	blocks := []notion.Block{
		notion.Block{
			Object: "block",
			Type:   "heading_3",
			Heading3: &notion.Text{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: header,
						},
					},
				},
			},
		},
		notion.Block{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &notion.TextTree{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: para,
						},
					},
				},
			},
		},
	}
	_, err := session.Client.AppendChildren(pageID, blocks)
	return err
}
