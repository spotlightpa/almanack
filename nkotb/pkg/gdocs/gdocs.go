// Package gdocs converts Google Documents to HTML.
package gdocs

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
)

func Request(ctx context.Context, cl *http.Client, docID string) (d *docs.Document, err error) {
	docID = NormalizeID(docID)
	var doc docs.Document
	if err = requests.
		URL("https://docs.googleapis.com").
		Pathf("/v1/documents/%s", docID).
		Client(cl).
		ToJSON(&doc).
		Fetch(ctx); err != nil {
		return nil, err
	}
	return &doc, nil
}

func NormalizeID(id string) string {
	if strings.Contains(id, "docID=") {
		if u, err := url.Parse(id); err == nil {
			id = u.Query().Get("docID")
		}
	}
	id = strings.TrimPrefix(id, "https://docs.google.com/document/d/")
	id, _, _ = strings.Cut(id, "/")
	return id
}

func Convert(doc *docs.Document) (n *html.Node) {
	n = &html.Node{
		Type: html.DocumentNode,
	}
	listInfo := buildListInfo(doc.Lists)
	objectInfo := buildObjectInfo(doc.InlineObjects)
	for _, el := range doc.Body.Content {
		convertEl(n, el, listInfo, objectInfo)
	}
	return
}

var tagForNamedStyle = map[string]string{
	"NAMED_STYLE_TYPE_UNSPECIFIED": "div",
	"NORMAL_TEXT":                  "p",
	"TITLE":                        "h1",
	"SUBTITLE":                     "h1",
	"HEADING_1":                    "h1",
	"HEADING_2":                    "h2",
	"HEADING_3":                    "h3",
	"HEADING_4":                    "h4",
	"HEADING_5":                    "h5",
	"HEADING_6":                    "h6",
}

func buildListInfo(lists map[string]docs.List) map[string]string {
	m := map[string]string{}
	for id, list := range lists {
		if list.ListProperties == nil {
			continue
		}
		listType := "ul"
		if list.ListProperties.NestingLevels[0].GlyphType != "" {
			listType = "ol"
		}
		m[id] = listType
	}
	return m
}

func buildObjectInfo(objs map[string]docs.InlineObject) map[string][]string {
	m := map[string][]string{}
	for id, obj := range objs {
		if obj.InlineObjectProperties == nil {
			continue
		}
		innerObj := obj.InlineObjectProperties.EmbeddedObject
		src := ""
		if innerObj.ImageProperties != nil {
			src = innerObj.ImageProperties.ContentUri
		}
		m[id] = []string{
			"src", src,
			"title", innerObj.Title,
			"alt", innerObj.Description,
		}
	}
	return m
}

func convertEl(n *html.Node, el *docs.StructuralElement, listInfo map[string]string, objInfo map[string][]string) {
	if el.Table != nil && el.Table.TableRows != nil {
		table := xhtml.New("table")
		n.AppendChild(table)
		for _, row := range el.Table.TableRows {
			rowEl := xhtml.New("tr")
			table.AppendChild(rowEl)
			if row.TableCells != nil {
				for _, cell := range row.TableCells {
					cellEl := xhtml.New("td")
					rowEl.AppendChild(cellEl)
					for _, content := range cell.Content {
						convertEl(cellEl, content, listInfo, objInfo)
					}
				}
			}
		}
	}
	if el.Paragraph == nil {
		return
	}
	if el.Paragraph.Bullet != nil {
		listType := listInfo[el.Paragraph.Bullet.ListId]
		ul := xhtml.LastChildOrNew(n, listType)
		li := xhtml.New("li")
		ul.AppendChild(li)
		n = li
	}

	blockType := tagForNamedStyle[el.Paragraph.ParagraphStyle.NamedStyleType]

	n.AppendChild(xhtml.New(blockType))

	for _, subel := range el.Paragraph.Elements {
		if subel.HorizontalRule != nil {
			n.AppendChild(xhtml.New("hr"))
		}

		if subel.InlineObjectElement != nil {
			inner := xhtml.LastChildOrNew(n, blockType)
			attrs := objInfo[subel.InlineObjectElement.InlineObjectId]
			inner.AppendChild(xhtml.New("img", attrs...))
		}

		if subel.TextRun == nil {
			continue
		}

		inner := xhtml.LastChildOrNew(n, blockType)
		if subel.TextRun.TextStyle != nil {
			if len(subel.TextRun.SuggestedInsertionIds) > 0 {
				newinner := xhtml.New("ins")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if len(subel.TextRun.SuggestedDeletionIds) > 0 {
				newinner := xhtml.New("del")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Link != nil {
				newinner := xhtml.New("a", "href", subel.TextRun.TextStyle.Link.Url)
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.BackgroundColor != nil {
				newinner := xhtml.New("mark")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Bold {
				newinner := xhtml.New("strong")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Italic {
				newinner := xhtml.New("em")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Underline && subel.TextRun.TextStyle.Link == nil {
				newinner := xhtml.New("u")
				inner.AppendChild(newinner)
				inner = newinner
			}
		}
		xhtml.AppendText(inner, subel.TextRun.Content)
	}
}
