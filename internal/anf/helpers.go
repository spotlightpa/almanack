package anf

import (
	"fmt"
	"time"

	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func generateIdentifier() string {
	// Generate a simple identifier based on timestamp
	return fmt.Sprintf("article-%d", time.Now().Unix())
}

func extractTitle(doc *html.Node) string {
	for n := range doc.Descendants() {
		if n.DataAtom == atom.Title {
			return xhtml.TextContent(n)
		}
	}
	return ""
}

// func extractTextWithFormatting(n *html.Node) (string, []InlineTextAddition) {
// 	var text strings.Builder
// 	var additions []InlineTextAddition

// 	extractFormattedText(n, &text, &additions)

// 	return strings.TrimSpace(text.String()), additions
// }

// func extractFormattedText(n *html.Node, text *strings.Builder, additions *[]InlineTextAddition) {
// 	if n.Type == html.TextNode {
// 		text.WriteString(n.Data)
// 		return
// 	}

// 	startPos := text.Len()

// 	// Process children first
// 	for child := range n.ChildNodes() {
// 		extractFormattedText(child, text, additions)
// 	}

// 	endPos := text.Len()
// 	length := endPos - startPos

// 	if length > 0 {
// 		switch n.DataAtom {
// 		case atom.Strong, atom.B:
// 			*additions = append(*additions, InlineTextAddition{
// 				Type:        "textStyle",
// 				TextStyle:   "bold",
// 				RangeStart:  startPos,
// 				RangeLength: length,
// 			})
// 		case atom.Em, atom.I:
// 			*additions = append(*additions, InlineTextAddition{
// 				Type:        "textStyle",
// 				TextStyle:   "italic",
// 				RangeStart:  startPos,
// 				RangeLength: length,
// 			})
// 		case atom.A:
// 			href := xhtml.Attr(n, "href")
// 			if href != "" {
// 				*additions = append(*additions, InlineTextAddition{
// 					Type:        "link",
// 					URL:         href,
// 					RangeStart:  startPos,
// 					RangeLength: length,
// 				})
// 			}
// 		}
// 	}
// }
