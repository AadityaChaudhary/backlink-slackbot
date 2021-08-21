package notion

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Client  *http.Client
	Token   string
	Version string
}

type Page struct {
	Id     *string `json:"page"`
	Object string  `json:"object"`
	Parent struct {
		Type string `json:"type"`
	} `json:"parent"`
	Properties *struct {
		Title struct {
			Id string `json:"id"`
			Type string `json:"type"`
			Title []RichText `json:"title"`
		} `json:"title"`
	} `json:"properties,omitempty"`
}

type Annotations struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

type TextInfo struct {
	Content string `json:"content"`
	Link    *struct {
		URL string `json:"url"`
	} `json:"link"`
}

type RichText struct {
	Type        string       `json:"type"`
	PlainText   *string      `json:"plain_text"`
	HREF        *string      `json:"href"`
	Annotations *Annotations `json:"annotations,omitempty"`

	Text *TextInfo `json:"text"`
}

type Text struct {
	Text []RichText `json:"text"`
}

type TextTree struct {
	Text     []RichText `json:"text"`
	Children []Block    `json:"children"`
}

type CheckedTextTree struct {
	Text     []RichText `json:"text"`
	Children []Block    `json:"children"`
	Checked  bool       `json:"checked"`
}

type Block struct {
	Id          *string `json:"id"`
	Object      string  `json:"object"`
	Type        string  `json:"type"`
	HasChildren bool    `json:"has_children"`

	Paragraph        *TextTree        `json:"paragraph"`
	Heading1         *Text            `json:"heading_1"`
	Heading2         *Text            `json:"heading_2"`
	Heading3         *Text            `json:"heading_3"`
	BulletedListItem *TextTree        `json:"bulleted_list_item"`
	NumberedListItem *TextTree        `json:"numbered_list_item"`
	ToDo             *CheckedTextTree `json:"to_do"`
	Toggle           *TextTree        `json:"toggle"`
	ChildPage        *struct {
		Title string `json:"title"`
	} `json:"child_page"`
}

type Database struct {
	Id     string `json:"id"`
	Object string `json:"object"`
	Parent struct {
		Type   string `json:"type"`
		PageId string `json:"page_id"`
	} `json:"parent"`

	Title []RichText `json:"title"`

	Properties struct {
		Description string `json:"description"`
	} `json:"properties"`
}

func QueryString(params map[string]string) string {
	var builder strings.Builder

	questionMark := true

	for key, value := range params {
		if questionMark {
			builder.WriteString("?")
			questionMark = false
		} else {
			builder.WriteString("&")
		}

		builder.WriteString(url.QueryEscape(key) + "=" + url.QueryEscape(value))
	}

	return builder.String()
}

func (client Client) MakeRequest(method string, path string, body string) ([]byte, error) {
	retries := 0
	var response *http.Response

	for {
		request, err := http.NewRequest(method, path, strings.NewReader(body))
		if err != nil {
			return []byte{}, err
		}

		request.Header.Set("Authorization", "Bearer "+client.Token)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Notion-Version", client.Version)

		response, err = client.Client.Do(request)
		if err != nil {
			return []byte{}, err
		}

		// Rate Limit? Try Again
		if response.StatusCode == 429 && retries < 3 {
			retries += 1
			time.Sleep(1000)
		} else {
			break
		}
	}

	out, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	if response.StatusCode != 200 {
		message := "status code " + strconv.Itoa(response.StatusCode) + ": " + string(out)
		return []byte{}, errors.New(message)
	}

	return out, nil
}

func (client Client) GetPage(id string) (Page, error) {
	path := "https://api.notion.com/v1/pages/" + id
	body, err := client.MakeRequest("GET", path, "")
	if err != nil {
		return Page{}, err
	}

	z := string(body)
	_ = z

	page := Page{}
	err = json.Unmarshal(body, &page)
	if err != nil {
		return Page{}, err
	}

	return page, nil
}

func (client Client) GetChildrenFrom(id string, from *string, size int) (BlockCursor, error) {
	path := "https://api.notion.com/v1/blocks/" + id + "/children"

	queries := map[string]string{
		"page_size": strconv.Itoa(size),
	}

	if from != nil {
		queries["start_cursor"] = *from
	}

	path += QueryString(queries)

	body, err := client.MakeRequest("GET", path, "")
	if err != nil {
		return BlockCursor{}, err
	}

	var data struct {
		Object     string
		Results    []Block
		NextCursor *string
		HasMore    bool
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return BlockCursor{}, err
	}
	if data.Object != "list" {
		return BlockCursor{}, errors.New("match issue")
	}

	next := func() (BlockCursor, error) {
		return client.GetChildrenFrom(id, data.NextCursor, size)
	}

	return BlockCursor{
		GetNext:   next,
		Current:   data.Results,
		Remaining: data.HasMore,
	}, nil
}

func (client Client) GetChildren(id string) (BlockCursor, error) {
	return client.GetChildrenFrom(id, nil, 50)
}

func (client Client) AppendChildren(id string, blocks []Block) (Block, error) {
	// First Block Only :think:
	value := struct {
		Children []Block `json:"children"`
	}{
		Children: blocks,
	}

	data, err := json.Marshal(value)
	if err != nil {
		return Block{}, err
	}

	// id -> parent block/page
	path := "https://api.notion.com/v1/blocks/" + id + "/children"
	body, err := client.MakeRequest("PATCH", path, string(data))
	if err != nil {
		return Block{}, err
	}

	var block Block
	err = json.Unmarshal(body, &block)
	if err != nil {
		return Block{}, err
	}
	if block.Object != "block" {
		return Block{}, errors.New("match issue")
	}

	return block, nil
}

func (client Client) GetDatabase(id string) (Database, error) {
	path := "https://api.notion.com/v1/databases/" + id

	body, err := client.MakeRequest("GET", path, "")
	if err != nil {
		return Database{}, err
	}

	var database Database
	err = json.Unmarshal(body, &database)
	if err != nil {
		return Database{}, err
	}

	return database, nil
}

func (client Client) GetDatabasePagesFrom(id string, from *string, size int) (PageCursor, error) {
	path := "https://api.notion.com/v1/databases/" + id + "/query"

	pages := struct {
		StartCursor *string `json:"start_cursor,omitempty"`
		PageSize    int     `json:"page_size"`
	}{
		StartCursor: from,
		PageSize:    size,
	}

	body, err := json.Marshal(pages)
	if err != nil {
		return PageCursor{}, err
	}

	response, err := client.MakeRequest("POST", path, string(body))
	if err != nil {
		return PageCursor{}, err
	}

	var info struct {
		Object     string
		Results    []Page
		NextCursor *string
		HasMore    bool
	}

	err = json.Unmarshal(response, &info)
	if err != nil {
		return PageCursor{}, err
	}
	if info.Object != "list" {
		return PageCursor{}, errors.New("match issue")
	}

	next := func() (PageCursor, error) {
		return client.GetDatabasePagesFrom(id, info.NextCursor, size)
	}

	return PageCursor{
		GetNext:   next,
		Current:   info.Results,
		Remaining: info.HasMore,
	}, nil
}

func (client Client) GetDatabasePages(id string) (PageCursor, error) {
	return client.GetDatabasePagesFrom(id, nil, 50)
}

// GetDatabasesFrom does not work
func (client Client) GetDatabasesFrom(from *string, size int) (DatabaseCursor, error) {
	path := "https://api.notion.com/v1/databases"

	queries := map[string]string{
		"page_size": strconv.Itoa(size),
	}

	if from != nil {
		queries["start_cursor"] = *from
	}

	path += QueryString(queries)

	data, err := client.MakeRequest("GET", path, "")
	if err != nil {
		return DatabaseCursor{}, err
	}

	var info struct {
		Object     string
		Results    []Database
		NextCursor *string
		HasMore    bool
	}

	err = json.Unmarshal(data, &info)
	if err != nil {
		return DatabaseCursor{}, err
	}
	if info.Object != "list" {
		return DatabaseCursor{}, errors.New("match issue")
	}

	next := func() (DatabaseCursor, error) {
		return client.GetDatabasesFrom(info.NextCursor, size)
	}

	return DatabaseCursor{
		GetNext:   next,
		Current:   info.Results,
		Remaining: info.HasMore,
	}, nil
}

// GetDatabases does not work
func (client Client) GetDatabases() (DatabaseCursor, error) {
	return client.GetDatabasesFrom(nil, 50)
}

func NewClient(token string) Client {
	return Client{
		Client:  &http.Client{},
		Token:   token,
		Version: "2021-05-13",
	}
}
