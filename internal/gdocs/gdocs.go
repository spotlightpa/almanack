// Package gdocs converts Google Documents to HTML.
package gdocs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/googleapi"
)

func Request(ctx context.Context, cl *http.Client, docID string) (d *docs.Document, err error) {
	var doc docs.Document
	type errorReply struct {
		Error googleapi.Error `json:"error"`
	}
	var errJSON errorReply
	if err = requests.
		URL("https://docs.googleapis.com").
		Pathf("/v1/documents/%s", docID).
		Param("suggestionsViewMode", "PREVIEW_WITHOUT_SUGGESTIONS").
		Client(cl).
		ErrorJSON(&errJSON).
		ToJSON(&doc).
		Fetch(ctx); err != nil {
		e := &errJSON.Error
		e.Wrap(err)
		return nil, resperr.WithCodeAndMessage(e, e.Code, e.Message)
	}
	return &doc, nil
}

// https://developers.google.com/docs/api/how-tos/overview#document_id
var validIDChars = bytemap.Union(
	bytemap.Range('A', 'Z'),
	bytemap.Range('a', 'z'),
	bytemap.Range('0', '9'),
	bytemap.Make("_-"),
)

func NormalizeID(id string) (string, error) {
	const magicLength = 44
	if len(id) == magicLength {
		return id, nil
	}
	var v resperr.Validator
	v.AddIf("gdocs_id", len(id) == 0, "ID must be set")
	rawID := id
	var found bool
	id, found = strings.CutPrefix(id, "https://docs.google.com/document/d/")
	v.AddIfUnset("gdocs_id", !found, "Unrecognized ID: %s", rawID)
	if found {
		id, _, _ = strings.Cut(id, "/")
	}
	v.AddIfUnset("gdocs_id", len(id) != magicLength, "Invalid ID: %s", rawID)
	v.AddIfUnset("gdocs_id", !validIDChars.Contains(id), "Illegal characters in ID: %s", rawID)
	return id, v.Err()
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
			"data-oid", obj.ObjectId,
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

		if person := subel.Person; person != nil {
			inner := xhtml.LastChildOrNew(n, blockType)
			link := xhtml.New("a", "href", fmt.Sprintf("mailto:%s", person.PersonProperties.Email))
			xhtml.AppendText(link, person.PersonProperties.Name)
			inner.AppendChild(link)
		}

		if richLink := subel.RichLink; richLink != nil {
			inner := xhtml.LastChildOrNew(n, blockType)
			link := xhtml.New("a", "href", richLink.RichLinkProperties.Uri)
			xhtml.AppendText(link, richLink.RichLinkProperties.Title)
			inner.AppendChild(link)
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
		// Remove PUA characters
		content := subel.TextRun.Content
		content = strings.ReplaceAll(content, "\uE907", "")
		xhtml.AppendText(inner, content)
	}
}
