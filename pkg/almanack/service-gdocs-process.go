package almanack

import (
	"cmp"
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"

	"github.com/earthboundkid/bytemap/v2"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// dataTagType is a string enum of possible types in
// <data type="" value=""> elements used to pass information
// from intermediate HTML to final processed documents
type dataTagType string

const (
	dtSpotlightRaw  dataTagType = "spl"
	dtSpotlightText dataTagType = "spl-text"
	dtPartnerText   dataTagType = "partner-text"
	dtDBEmbed       dataTagType = "db-embed"
)

func newDataTag(dtype dataTagType, value string) *html.Node {
	data := xhtml.New("data", "type", string(dtype), "value", value)
	return data
}

func dbEmbedFromString(s string) db.Embed {
	var dbembed db.Embed
	must.Do(json.Unmarshal([]byte(s), &dbembed))
	return dbembed
}

func dbEmbedToString(embed db.Embed) string {
	return string(must.Get(json.Marshal(embed)))
}

// dataEls yields a tuple of elements and their value attribute
// for data elements with a matching type="" attribute.
func dataEls(n *html.Node, tag dataTagType) iter.Seq2[*html.Node, string] {
	return func(yield func(*html.Node, string) bool) {
		els := xhtml.SelectSlice(n, func(n *html.Node) bool {
			return n.DataAtom == atom.Data && xhtml.Attr(n, "type") == string(tag)
		})

		for _, el := range els {
			if !yield(el, xhtml.Attr(el, "value")) {
				return
			}
		}
	}
}

var ascii = bytemap.Range(0, 127)

func processDocHTML(docHTML *html.Node) (
	metadata db.GDocsMetadata,
	embeds []db.Embed,
	richText *html.Node, rawHTML *html.Node,
	markdown string,
	warnings []string,
) {
	// Now collect the embeds array and metadata
	n := 1
	for tbl, rows := range xhtml.Tables(docHTML) {
		embed := db.Embed{N: n}
		switch label := rows.Label(); label {
		case "html", "embed", "raw", "script":
			embed.Type = db.RawEmbedTag
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			embed.Value = embedHTML
			if !ascii.Contains(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d contains unusual characters.", n,
				))
			}
			if !xhtml.IsBalanced(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d seems to contain unbalanced HTML.", n,
				))
			}
			goto append

		case "spl", "spl-embed":
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			if !strings.Contains(embedHTML, "{{<") && !xhtml.IsBalanced(embedHTML) {
				warnings = append(warnings,
					"Spotlight PA embed seems to contain unbalanced HTML.")
			}
			data := newDataTag(dtSpotlightRaw, embedHTML)
			xhtml.ReplaceWith(tbl, data)

		case "spl-text":
			n := xhtml.Clone(rows.At(1, 0))
			blocko.MergeSiblings(n)
			blocko.RemoveEmptyP(n)
			blocko.RemoveMarks(n)
			s := blocko.Blockize(n)
			data := newDataTag(dtSpotlightText, s)
			xhtml.ReplaceWith(tbl, data)

		case "partner-embed":
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			embed.Type = db.PartnerRawEmbedTag
			embed.Value = embedHTML
			if !ascii.Contains(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d contains unusual characters.", n,
				))
			}
			if !xhtml.IsBalanced(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d seems to contain unbalanced HTML.", n,
				))
			}
			goto append

		case "partner-text":
			n := xhtml.Clone(rows.At(1, 0))
			blocko.MergeSiblings(n)
			blocko.RemoveEmptyP(n)
			blocko.RemoveMarks(n)
			data := newDataTag(dtPartnerText, xhtml.InnerHTMLBlocks(n))
			xhtml.ReplaceWith(tbl, data)

		case "photo", "image", "photograph", "illustration", "illo":
			embed.Type = db.ImageEmbedTag
			if imageEmbed, warning := processImage(rows, n); warning != "" {
				tbl.Parent.RemoveChild(tbl)
				warnings = append(warnings, warning)
			} else {
				embed.Value = *imageEmbed
				goto append
			}

		case "metadata", "info":
			processMetadata(rows, &metadata)
			tbl.Parent.RemoveChild(tbl)

		case "comment", "ignore", "note":
			tbl.Parent.RemoveChild(tbl)

		case "table":
			row := xhtml.Closest(rows.At(0, 0), xhtml.WithAtom(atom.Tr))
			row.Parent.RemoveChild(row)

		case "toc", "table of contents":
			embed.Type = db.ToCEmbedTag
			embed.Value = processToc(docHTML, rows)
			goto append

		default:
			warnings = append(warnings, fmt.Sprintf(
				"Unrecognized table type: %q", label,
			))
			tbl.Parent.RemoveChild(tbl)
		}
		continue
	append:
		embeds = append(embeds, embed)
		data := newDataTag(dtDBEmbed, dbEmbedToString(embed))
		xhtml.ReplaceWith(tbl, data)
		n++
	}

	docHTML = must.Get(blocko.Minify(xhtml.ToBuffer(docHTML)))

	blocko.MergeSiblings(docHTML)
	blocko.RemoveEmptyP(docHTML)
	blocko.RemoveMarks(docHTML)

	// Warn about fake headings
	for n := range xhtml.ChildNodes(docHTML) {
		// <p> with only b/i/strong/em for a child
		if n.DataAtom != atom.P {
			continue
		}
		if n.FirstChild != nil &&
			n.FirstChild == n.LastChild &&
			slices.Contains([]atom.Atom{
				atom.B, atom.Strong,
			}, n.FirstChild.DataAtom) {
			text := xhtml.TextContent(n)
			if len([]rune(text)) > 17 {
				runes := []rune(text)[:13]
				text = string(runes) + "..."
			}
			warning := fmt.Sprintf(
				"Paragraph beginning %q looks like a header, but does not use H-tag.", text)
			warnings = append(warnings, warning)
		}
	}

	// Warn about <br>
	if n := xhtml.Select(docHTML, xhtml.WithAtom(atom.Br)); n != nil {
		warnings = append(warnings,
			"Document contains <br> line breaks. Are you sure you want to use a line break? In Google Docs, select View > Show non-printing characters to see them.")
	}

	// Clone turn data elements into partner placeholders
	richText = xhtml.Clone(docHTML)
	intermediateDocToPartnerRichText(richText)

	// For rawHTML, convert to raw nodes
	rawHTML = xhtml.Clone(docHTML)
	intermediateDocToPartnerHTML(rawHTML)

	// Markdown data conversion
	markdown = intermediateDocToMarkdown(docHTML)

	return
}

func processImage(rows xhtml.TableNodes, n int) (imageEmbed *db.EmbedImage, warning string) {
	var width, height int
	if w := xhtml.TextContent(rows.Value("width")); w != "" {
		width, _ = strconv.Atoi(w)
	}
	if h := xhtml.TextContent(rows.Value("height")); h != "" {
		height, _ = strconv.Atoi(h)
	}
	imageEmbed = &db.EmbedImage{
		Credit:  xhtml.TextContent(rows.Value("credit")),
		Caption: xhtml.TextContent(rows.Value("caption")),
		Description: cmp.Or(
			xhtml.TextContent(rows.Value("description")),
			xhtml.TextContent(rows.Value("alt")),
		),
		Width:  width,
		Height: height,
	}

	if path := xhtml.TextContent(rows.Value("path")); path != "" {
		imageEmbed.Path = path
		return imageEmbed, ""
	}
	return nil, fmt.Sprintf(
		"Table %d missing image", n,
	)
}

func processMetadata(rows xhtml.TableNodes, metadata *db.GDocsMetadata) {
	metadata.InternalID = cmp.Or(
		xhtml.TextContent(rows.Value("slug")),
		xhtml.TextContent(rows.Value("internal id")),
		metadata.InternalID,
	)
	metadata.Byline = cmp.Or(
		xhtml.TextContent(rows.Value("byline")),
		xhtml.TextContent(rows.Value("authors")),
		xhtml.TextContent(rows.Value("author")),
		xhtml.TextContent(rows.Value("by")),
	)
	if strings.HasPrefix(metadata.Byline, "By ") ||
		strings.HasPrefix(metadata.Byline, "by ") {
		metadata.Byline = metadata.Byline[3:]
	}
	metadata.Budget = xhtml.TextContent(rows.Value("budget"))
	metadata.Hed = cmp.Or(
		xhtml.TextContent(rows.Value("hed")),
		xhtml.TextContent(rows.Value("title")),
		xhtml.TextContent(rows.Value("headline")),
		xhtml.TextContent(rows.Value("hedline")),
	)
	metadata.Description = cmp.Or(
		xhtml.TextContent(rows.Value("seo description")),
		xhtml.TextContent(rows.Value("description")),
		xhtml.TextContent(rows.Value("desc")),
	)
	metadata.LedeImageCredit = cmp.Or(
		xhtml.TextContent(rows.Value("lede image credit")),
		xhtml.TextContent(rows.Value("lead image credit")),
		xhtml.TextContent(rows.Value("credit")),
	)
	metadata.LedeImageCaption = cmp.Or(
		xhtml.TextContent(rows.Value("lede image caption")),
		xhtml.TextContent(rows.Value("lead image caption")),
		xhtml.TextContent(rows.Value("caption")),
	)
	metadata.LedeImageDescription = cmp.Or(
		xhtml.TextContent(rows.Value("lede image description")),
		xhtml.TextContent(rows.Value("lead image description")),
		xhtml.TextContent(rows.Value("lede image alt")),
		xhtml.TextContent(rows.Value("lead image alt")),
		xhtml.TextContent(rows.Value("alt")),
	)
	metadata.URLSlug = cmp.Or(
		xhtml.TextContent(rows.Value("url")),
		xhtml.TextContent(rows.Value("keywords")),
	)
	metadata.URLSlug = strings.TrimRight(metadata.URLSlug, "/")
	_, metadata.URLSlug, _ = stringx.LastCut(metadata.URLSlug, "/")
	metadata.URLSlug = stringx.SlugifyURL(metadata.URLSlug)

	metadata.Blurb = cmp.Or(
		xhtml.TextContent(rows.Value("blurb")),
		xhtml.TextContent(rows.Value("summary")),
	)
	metadata.LinkTitle = cmp.Or(
		xhtml.TextContent(rows.Value("link title")),
	)
	metadata.SEOTitle = cmp.Or(
		xhtml.TextContent(rows.Value("seo hed")),
		xhtml.TextContent(rows.Value("seo title")),
		xhtml.TextContent(rows.Value("seo headline")),
		xhtml.TextContent(rows.Value("seo hedline")),
	)
	metadata.OGTitle = cmp.Or(
		xhtml.TextContent(rows.Value("facebook hed")),
		xhtml.TextContent(rows.Value("facebook title")),
	)
	metadata.TwitterTitle = cmp.Or(
		xhtml.TextContent(rows.Value("twitter hed")),
		xhtml.TextContent(rows.Value("twitter title")),
	)
	metadata.Eyebrow = cmp.Or(
		xhtml.TextContent(rows.Value("eyebrow")),
		xhtml.TextContent(rows.Value("kicker")),
	)

	metadata.LedeImage = cmp.Or(
		xhtml.TextContent(rows.Value("lede image path")),
		xhtml.TextContent(rows.Value("lead image path")),
		xhtml.TextContent(rows.Value("path")),
	)
}

func processToc(doc *html.Node, rows xhtml.TableNodes) string {
	type header struct {
		text  string
		id    string
		depth int
	}
	var headers []header
	for n := range xhtml.All(doc) {
		switch n.DataAtom {
		case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		default:
			continue
		}
		id := fmt.Sprintf("spl-heading-%d", len(headers)+1)
		xhtml.SetAttr(n, "id", id)
		depth := int(n.Data[1] - '0')
		headers = append(headers, header{xhtml.TextContent(n), id, depth})
	}
	container := xhtml.New("div")
	h3 := xhtml.New("h3")
	xhtml.AppendText(h3, cmp.Or(
		xhtml.TextContent(rows.At(0, 1)),
		xhtml.TextContent(rows.At(1, 0)),
		"Table of Contents",
	))
	container.AppendChild(h3)
	ul := xhtml.New("ul")
	container.AppendChild(ul)
	currentUl := ul
	lastDepth := 7 // Past H6, the maximum possible depth
	for _, h := range headers {
		// If this one is deeper or less deep than its predecessor,
		// add and remove ULs as needed
		d := h.depth
		for lastDepth > d {
			// If its out of order, just try to cope
			currentUl = cmp.Or(
				xhtml.Closest(currentUl.Parent, xhtml.WithAtom(atom.Ul)),
				currentUl,
			)
			d++
		}
		for lastDepth < d {
			newUl := xhtml.New("ul")
			lastLi := xhtml.LastChildOrNew(currentUl, "li")
			lastLi.AppendChild(newUl)
			currentUl = newUl
			d--
		}
		li := xhtml.New("li")
		p := xhtml.New("p")
		link := xhtml.New("a", "href", "#"+h.id)
		xhtml.AppendText(link, h.text)
		p.AppendChild(link)
		li.AppendChild(p)
		currentUl.AppendChild(li)
		lastDepth = h.depth
	}

	return xhtml.InnerHTML(container)
}
