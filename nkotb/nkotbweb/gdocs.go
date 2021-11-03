package nkotbweb

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
)

func getDoc(ctx context.Context, cl *http.Client, docID string) (n *html.Node, err error) {
	docID = normalizeID(docID)
	var doc docs.Document
	if err = requests.
		URL("https://docs.googleapis.com").
		Pathf("/v1/documents/%s", docID).
		Client(cl).
		ToJSON(&doc).
		Fetch(ctx); err != nil {
		return nil, err
	}

	n = convert(&doc)
	return n, nil
}

func normalizeID(id string) string {
	id = strings.TrimPrefix(id, "https://docs.google.com/document/d/")
	id, _, _ = cut(id, "/")
	return id
}

func convert(doc *docs.Document) (n *html.Node) {
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
		table := newElement("table")
		n.AppendChild(table)
		for _, row := range el.Table.TableRows {
			rowEl := newElement("tr")
			table.AppendChild(rowEl)
			if row.TableCells != nil {
				for _, cell := range row.TableCells {
					cellEl := newElement("td")
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
		ul := lastChildOrNewElement(n, listType)
		li := newElement("li")
		ul.AppendChild(li)
		n = li
	}

	blockType := tagForNamedStyle[el.Paragraph.ParagraphStyle.NamedStyleType]

	n.AppendChild(newElement(blockType))

	for _, subel := range el.Paragraph.Elements {
		if subel.HorizontalRule != nil {
			n.AppendChild(newElement("hr"))
		}

		if subel.InlineObjectElement != nil {
			inner := lastChildOrNewElement(n, blockType)
			attrs := objInfo[subel.InlineObjectElement.InlineObjectId]
			inner.AppendChild(newElement("img", attrs...))
		}

		if subel.TextRun == nil {
			continue
		}

		inner := lastChildOrNewElement(n, blockType)
		if subel.TextRun.TextStyle != nil {
			if len(subel.TextRun.SuggestedInsertionIds) > 0 {
				newinner := newElement("ins")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if len(subel.TextRun.SuggestedDeletionIds) > 0 {
				newinner := newElement("del")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Link != nil {
				newinner := newElement("a", "href", subel.TextRun.TextStyle.Link.Url)
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.BackgroundColor != nil {
				newinner := newElement("mark")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Bold {
				newinner := newElement("strong")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Italic {
				newinner := newElement("em")
				inner.AppendChild(newinner)
				inner = newinner
			}
			if subel.TextRun.TextStyle.Underline && subel.TextRun.TextStyle.Link == nil {
				newinner := newElement("u")
				inner.AppendChild(newinner)
				inner = newinner
			}
		}
		appendText(inner, subel.TextRun.Content)
	}
}

func newElement(tag string, attrs ...string) *html.Node {
	var attrslice []html.Attribute
	if len(attrs) > 0 {
		if len(attrs)%2 != 0 {
			panic("uneven number of attr/value pairs")
		}
		attrslice = make([]html.Attribute, len(attrs)/2)
		for i := range attrslice {
			attrslice[i] = html.Attribute{
				Key: attrs[i*2],
				Val: attrs[i*2+1],
			}
		}
	}
	return &html.Node{
		Type: html.ElementNode,
		Data: tag,
		Attr: attrslice,
	}
}

func lastChildOrNewElement(p *html.Node, tag string, attrs ...string) *html.Node {
	if p.LastChild != nil && p.LastChild.Data == tag {
		return p.LastChild
	}
	n := newElement(tag, attrs...)
	p.AppendChild(n)
	return n
}

func appendText(n *html.Node, text string) {
	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}

var scopes = []string{
	"https://www.googleapis.com/auth/documents",
	"https://www.googleapis.com/auth/documents.readonly",
	"https://www.googleapis.com/auth/drive",
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/drive.readonly",
}

func makeStateToken() (string, error) {
	var b [15]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b[:]), nil
}

// See https://github.com/golang/go/issues/46336
func cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}
