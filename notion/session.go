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
	case "Paragraph": return block.Paragraph.Text
	case "Heading1": return block.Heading1.Text
	case "Heading2": return block.Heading2.Text
	case "Heading3": return block.Heading3.Text
	case "BulletedListItem": return block.BulletedListItem.Text
	case "NumberedListItem": return block.NumberedListItem.Text
	case "ToDo": return block.ToDo.Text
	case "Toggle": return block.Toggle.Text
	}

	return nil
}

func (block Block) GetChildren() []Block {
	switch block.Type {
	case "Paragraph": return block.Paragraph.Children
	case "BulletedListItem": return block.BulletedListItem.Children
	case "NumberedListItem": return block.NumberedListItem.Children
	case "ToDo": return block.ToDo.Children
	case "Toggle": return block.Toggle.Children
	}

	return nil
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

func ParseSections(client Client, blocks []Block) []InterfaceSection {
	var sections []InterfaceSection

	var index = 0

	var section *InterfaceSection

	predicate := func(name string) bool {
		heading := "heading"

		return name[0:len(heading)] == heading
	}

	for index < len(blocks) {
		block := &blocks[index]

		if predicate(block.Type) || section == nil {
			if section != nil {
				sections = append(sections, *section)
			}

			section = &InterfaceSection {
				Client: client,
				Head: block,

				Title: Flatten(block.GetText()),
				Elements: []InterfaceElement { },
			}
		} else {
			section.Elements = append(section.Elements, InterfaceElement {
				Text: Flatten(block.GetText()),
				Block: block,
			})
		}

		index++
	}

	if section != nil {
		sections = append(sections, *section)
	}

	return sections
}

func (page *InterfacePage) Reload() error {
	value, err := page.Client.GetPage(page.Id)
	if err != nil { return err }

	page.Title = Flatten(value.Properties.Title.Title)

	cursor, err := page.Client.GetChildren(page.Id)
	if err != nil { return err }

	blocks := cursor.ReadAll()

	page.Sections = ParseSections(page.Client, blocks)

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
