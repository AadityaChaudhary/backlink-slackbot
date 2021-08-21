package notion

import (
	"strings"
)

type InterfaceElement struct {
	Text string
	Block *Block
}

type InterfaceSection struct {
	Client Client

	Head *Block

	Title string
	Elements []InterfaceElement
}

type InterfacePage struct {
	Client Client

	Id string

	Title string
	Sections []InterfaceSection
}

type Session struct {
	Client Client
	Pages []InterfacePage
}

func (block Block) GetText() []RichText {
	switch block.Type {
	case "paragraph": return block.Paragraph.Text
	case "heading_1": return block.Heading1.Text
	case "heading_2": return block.Heading2.Text
	case "heading_3": return block.Heading3.Text
	case "bulleted_list_item": return block.BulletedListItem.Text
	case "numbered_list_item": return block.NumberedListItem.Text
	case "to_do": return block.ToDo.Text
	case "toggle": return block.Toggle.Text
	}

	return nil
}

func (block Block) TypeHasChildren() bool {
	switch block.Type {
	case "paragraph": return true
	case "bulleted_list_item": return true
	case "numbered_list_item": return true
	case "to_do": return true
	case "toggle": return true
	}

	return false
}

func Flatten(objects []RichText) string {
	var builder strings.Builder

	for _, text := range objects {
		if text.PlainText != nil {
			builder.WriteString(*text.PlainText)
		}
	}

	return builder.String()
}

func ParseSections(client Client, blocks []Block) ([]InterfaceSection, error) {
	var sections []InterfaceSection

	var index = 0

	for index < len(blocks) {
		block := &blocks[index]

		if block.TypeHasChildren() {
			var elements []InterfaceElement

			if block.HasChildren {
				cursor, err := client.GetChildren(*block.Id)
				if err != nil { return nil, err }

				children := cursor.ReadAll()

				for _, element := range children {
					text := element.GetText()

					var flattened string

					if text != nil {
						flattened = Flatten(text)
					}

					elements = append(elements, InterfaceElement{
						Text:  flattened,
						Block: &element,
					})
				}
			}

			sections = append(sections, InterfaceSection {
				Client: client,
				Head: block,

				Title: Flatten(block.GetText()),
				Elements: elements,
			})
		}

		index++
	}

	return sections, nil
}

func (page *InterfacePage) Reload() error {
	value, err := page.Client.GetPage(page.Id)
	if err != nil { return err }

	page.Title = Flatten(value.Properties.Title.Title)

	cursor, err := page.Client.GetChildren(page.Id)
	if err != nil { return err }

	blocks := cursor.ReadAll()

	sections, err := ParseSections(page.Client, blocks)
	if err != nil { return err }

	page.Sections = sections

	return nil
}

func NewSession(client Client, pageIds []string) (Session, error) {
	var pages []InterfacePage

	for _, id := range pageIds {
		page := InterfacePage {
			Client: client,
			Id: id,
		}

		err := page.Reload()
		if err != nil { return Session {}, err }

		pages = append(pages, page)
	}

	session := Session {
		Client: client,
		Pages: pages,
	}

	return session, nil
}
