package slack

import (
	"backlink/db"
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
	var t string
	var link string
	var err error
	backlinks := getBacklinks(ev.Text)
	if len(backlinks) == 0 {
		log.Println("no backlinks found")
		return
	}
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
		t = msgs[0].Timestamp

		link, err = api.GetPermalink(&slack.PermalinkParameters{Channel: ev.Channel, Ts: msgs[0].Timestamp})
		if err != nil {
			log.Println("err", err)
			return
		}
	} else {

		txt = ev.Text
		userID = ev.User
		t = ev.TimeStamp
		link, err = api.GetPermalink(&slack.PermalinkParameters{Channel: ev.Channel, Ts: ev.TimeStamp})
		if err != nil {
			log.Println("err", err)
			return
		}
	}

	u, err := api.GetUserInfoContext(context.Background(), userID)
	if err != nil {
		log.Println(err)
		return
	}
	user = u.Profile.RealName
	timeS, err := convertTime(t)
	log.Println(backlinks)
	log.Println(txt)
	log.Println(user)
	log.Println(t)
	log.Println(timeS.UTC())

	teamName, err := GetTeamName(api)
	if err != nil {
		log.Println(err)
		return
	}
	header := fmt.Sprint(user, " ", timeS.Format(time.RFC822))

	for _, backlink := range backlinks {
		if db.BacklinkExists(teamName, backlink) {
			pID, err := db.GetNotionID(teamName, backlink)
			if err != nil {
				log.Println("b", backlink, "err", err)
				return
			}
			err = addContent(session, pID, header, txt, link)
			if err != nil {
				log.Println("b", backlink, "err", err)
				return
			}
		} else {
			pID, err := createNewBacklinkPage(session, backlink, header, txt, link)
			if err != nil {
				log.Println("b", backlink, "err", err)
				return
			}
			bldb := db.Backlink{LinkName: backlink, NotionID: pID}
			db.AddBacklinkToWorkspace(teamName, bldb)
		}

	}

	//	_, err = createNewBacklinkPage(session, backlinks[0], header, txt)
	//	if err != nil {
	//		log.Println("err", err)
	//		return
	//	}

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

func createNewBacklinkPage(session *notion.Session, title, header, para, link string) (string, error) {

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
		notion.Block{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &notion.TextTree{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: "Go To Message",
							Link: &notion.Link{
								URL: link,
							},
						},
					},
				},
			},
		},
	}
	p, err := session.Pages[0].AppendPageWithBlocks(title, blocks)
	return p.Id, err
}

func addContent(session *notion.Session, pageID, header, para, link string) error {
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
		notion.Block{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &notion.TextTree{
				Text: []notion.RichText{
					notion.RichText{
						Type: "text",
						Text: &notion.TextInfo{
							Content: "Go To Message",
							Link: &notion.Link{
								URL: link,
							},
						},
					},
				},
			},
		},
	}
	_, err := session.Client.AppendChildren(pageID, blocks)
	return err
}
