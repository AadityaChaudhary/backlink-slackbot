package notion

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	Client *http.Client
	Token string
	Version string
}

type Page struct {
	Id *string `json:"page"`
	Object string `json:"object"`
	Parent struct {
		Type string `json:"type"`
	}  `json:"parent"`
	Properties map[string]interface{} `json:"properties"`
}

type Annotations struct {
	Bold bool `json:"bold"`
	Italic bool `json:"italic"`
	Strikethrough bool `json:"strikethrough"`
	Underline bool `json:"underline"`
	Code bool `json:"code"`
	Color string `json:"color"`
}

type NotionTextInfo struct {
	Content string `json:"content"`
	Link *struct {
		URL string `json:"url"`
	} `json:"link"`
}

type RichText struct {
	Type string `json:"type"`
	PlainText *string `json:"plain_text"`
	HREF *string             `json:"href"`
	Annotations *Annotations `json:"annotations,omitempty"`

	Text *NotionTextInfo `json:"text"`
}

type Text struct {
	Text []RichText `json:"text"`
}

type TextTree struct {
	Text []RichText  `json:"text"`
	Children []Block `json:"children"`
}

type CheckedTextTree struct {
	Text []RichText  `json:"text"`
	Children []Block `json:"children"`
	Checked bool     `json:"checked"`
}

type Block struct {
	Id *string `json:"id"`
	Object string `json:"object"`
	Type string `json:"type"`
	HasChildren bool `json:"has_children"`

	Paragraph *TextTree         `json:"paragraph"`
	Heading1 *Text              `json:"heading_1"`
	Heading2 *Text                   `json:"heading_2"`
	Heading3 *Text              `json:"heading_3"`
	BulletedListItem *TextTree  `json:"bulleted_list_item"`
	NumberedListItem *TextTree `json:"numbered_list_item"`
	ToDo *CheckedTextTree      `json:"to_do"`
	Toggle *TextTree           `json:"toggle"`
	ChildPage *struct {
		Title string `json:"title"`
	} `json:"child_page"`
}

type BlockCursor struct {
	GetNext func(string)(BlockCursor, error)
	Current []Block

	Cursor *string
}

func (cursor *BlockCursor) Next() error {
	if cursor.Cursor == nil {
		cursor.Current = []Block{ }
		return errors.New("end of list")
	}

	next, err := cursor.GetNext(*cursor.Cursor)
	if err != nil { return err }

	*cursor = next

	return nil
}

func (cursor *BlockCursor) ReadAll() []Block {
	all := cursor.Current

	for cursor.Cursor != nil {
		err := cursor.Next()
		if err != nil { return all }

		all = append(all, cursor.Current...)
	}

	return all
}

func (client Client) MakeRequest(method string, path string, body string) ([]byte, error) {
	request, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil { return []byte {}, err }

	request.Header.Set("Authorization", "Bearer " + client.Token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Notion-Version", client.Version)

	response, err := client.Client.Do(request)
	if err != nil { return []byte {}, err }

	out, err := ioutil.ReadAll(response.Body)
	if err != nil { return []byte {}, err }

	if response.StatusCode != 200 {
		message := "status code " + strconv.Itoa(response.StatusCode) + ": " + string(out)
		return []byte {}, errors.New(message)
	}

	return out, nil
}

func (client Client) GetPage(id string) (Page, error) {
	path := "https://api.notion.com/v1/pages/" + id
	body, err := client.MakeRequest("GET", path, "")
	if err != nil { return Page{}, err }

	page := Page{ }
	err = json.Unmarshal(body, &page)
	if err != nil { return Page{}, err }

	return page, nil
}

func (client Client) GetChildrenFrom(id string, from *string, size int) (BlockCursor, error) {
	path := "https://api.notion.com/v1/blocks/" + id + "/children"

	questionMark := false

	if from != nil {
		// this is very bad... injection could come from anywhere :|
		// but I haven't found any builtin go tool to escape query strings...
		path += "?start_cursor=" + url.QueryEscape(*from)
		questionMark = true
	}

	// wwwww pain hack
	if questionMark {
		path += "&"
	} else {
		path += "&"
	}

	path += "page_size=" + url.QueryEscape(strconv.Itoa(size))

	body, err := client.MakeRequest("GET", path, "")
	if err != nil { return BlockCursor{}, err }

	var data struct {
		Object string
		Results []Block
		NextCursor *string
		HasMore bool
	}
	err = json.Unmarshal(body, &data)
	if err != nil { return BlockCursor{}, err }
	if data.Object != "list" { return BlockCursor{}, errors.New("match issue") }

	next := func(cursor string) (BlockCursor, error) {
		return client.GetChildrenFrom(id, &cursor, size)
	}

	return BlockCursor{
		GetNext: next,
		Current: data.Results,
		Cursor: data.NextCursor,
	}, nil
}

func (client Client) GetChildren(id string) (BlockCursor, error) {
	return client.GetChildrenFrom(id, nil, 50)
}

func (client Client) AppendChildren(id string, blocks []Block) (Block, error) {
	// First Block Only :think:
	value := struct {
		Children []Block `json:"children"`
	} {
		Children: blocks,
	}

	data, err := json.Marshal(value)
	if err != nil { return Block{}, err }

	// id -> parent block/page
	path := "https://api.notion.com/v1/blocks/" + id + "/children"
	body, err := client.MakeRequest("PATCH", path, string(data))
	if err != nil { return Block{}, err }

	var block Block
	err = json.Unmarshal(body, &block)
	if err != nil { return Block{}, err }
	if block.Object != "block" { return Block{}, errors.New("match issue") }

	return block, nil
}

func NewClient(token string) Client {
	return Client{
		Client: &http.Client { },
		Token: token,
		Version: "2021-05-13",
	}
}
