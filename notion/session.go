package notion

import (
	"strings"
)

type InterfaceElement struct {
	Text string
	Block Block
}

type InterfaceSection struct {
	Client Client

	Head Block

	Title string
	Elements []InterfaceElement
}

type InterfacePage struct {
	Client Client

	Id string

	Title string
	Children []InterfacePage
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

func ParseSections(client Client, blocks []Block) ([]InterfaceSection, []InterfacePage, error) {
	var children []InterfacePage
	var sections []InterfaceSection

	var index = 0

	for index < len(blocks) {
		block := blocks[index]

		if block.Type == "child_page" {
			page := InterfacePage {
				Client: client,
				Id: *block.Id,
			}

			err := page.Reload()
			if err != nil { return nil, nil, err }

			children = append(children, page)
		} else if block.TypeHasChildren() {
			var elements []InterfaceElement

			if block.HasChildren {
				cursor, err := client.GetChildren(*block.Id)
				if err != nil { return nil, nil, err }

				children := cursor.ReadAll()

				for _, element := range children {
					text := element.GetText()

					var flattened string

					if text != nil {
						flattened = Flatten(text)
					}

					elements = append(elements, InterfaceElement{
						Text:  flattened,
						Block: element,
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

	return sections, children, nil
}

func (section *InterfaceSection) AppendBlock(block Block) (InterfaceElement, error) {
	value, err := section.Client.AppendChildren(*section.Head.Id, []Block { block })
	if err != nil { return InterfaceElement {}, err }

	element := InterfaceElement {
		Text: Flatten(value.GetText()),
		Block: value,
	}

	section.Elements = append(section.Elements, element)

	return element, nil
}

func (section *InterfaceSection) AppendElement(description string) (InterfaceElement, error) {
	block := Block {
		Object: "block",
		Type:   "paragraph",
		Paragraph: &TextTree {
			Text: []RichText {
				{
					Type: "text",
					Text: &TextInfo {
						Content: description,
					},
				},
			},
		},
	}

	return section.AppendBlock(block)
}

func (page *InterfacePage) AppendSection(description string, heading string) (InterfaceSection, error) {
	// Append Children only returns one block... makes me feel like I have to insert them separately

	if len(heading) > 0 {
		headingBlock := Block {
			Object: "block",
			Type:   "heading_1",
			Heading1: &Text {
				Text: []RichText {
					{
						Type: "text",
						Text: &TextInfo {
							Content: heading,
						},
					},
				},
			},
		}

		_, err := page.Client.AppendChildren(page.Id, []Block { headingBlock })
		if err != nil { return InterfaceSection {}, err }
	}

	startBlock := Block {
		Object: "block",
		Type:   "paragraph",
		Paragraph: &TextTree {
			Text: []RichText {
				{
					Type: "text",
					Text: &TextInfo {
						Content: description,
					},
				},
			},
		},
	}

	block, err := page.Client.AppendChildren(page.Id, []Block { startBlock })
	if err != nil { return InterfaceSection {}, err }

	section := InterfaceSection {
		Client: page.Client,

		Head: block, // Might not have a lot of data :)
		Title: Flatten(block.GetText()),
		Elements: nil, // no sub elements yet hopefully D:
	}

	page.Sections = append(page.Sections, section)

	return section, err
}

func (page *InterfacePage) AppendPageWithBlocks(title string, blocks []Block) (InterfacePage, error) {
	value, err := page.Client.CreatePageWithBlocks(page.Id, title, blocks)
	if err != nil { return InterfacePage {}, err }

	result := InterfacePage {
		Client: page.Client,

		Id: *value.Id,
		Title: Flatten(value.Properties.Title.Title),
	}

	page.Children = append(page.Children, result)

	return result, nil
}

func (page *InterfacePage) AppendPage(title string) (InterfacePage, error) {
	return page.AppendPageWithBlocks(title, []Block { })
}

func (page *InterfacePage) Reload() error {
	value, err := page.Client.GetPage(page.Id)
	if err != nil { return err }

	page.Title = Flatten(value.Properties.Title.Title)

	cursor, err := page.Client.GetChildren(page.Id)
	if err != nil { return err }

	blocks := cursor.ReadAll()

	sections, children, err := ParseSections(page.Client, blocks)
	if err != nil { return err }

	page.Sections = sections
	page.Children = children

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
