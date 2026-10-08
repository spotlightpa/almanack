package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/google"
	docs "google.golang.org/api/docs/v1"
)

// minimalJPEG is a valid JPEG image padded to >512 bytes so that
// the MIME sniffer can detect it as image/jpeg.
var minimalJPEG = func() []byte {
	// SOI + JFIF APP0 header + EOI, padded with zero bytes
	header := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, 0x00, 0x10, // APP0 marker, length 16
		0x4A, 0x46, 0x49, 0x46, 0x00, // 'JFIF\0'
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // version, density, thumbnail
		0xFF, 0xD9, // EOI
	}
	return append(header, make([]byte, 600)...)
}()

// minimalJPEGResponse is a raw HTTP/1.1 response serving minimalJPEG.
var minimalJPEGResponse = fmt.Sprintf(
	"HTTP/1.1 200 OK\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n%s",
	len(minimalJPEG),
	string(minimalJPEG),
)

// normalParagraphStyle is the minimum ParagraphStyle needed to avoid nil
// dereferences in gdocs.Convert (it accesses ParagraphStyle.NamedStyleType).
var normalParagraphStyle = &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"}

// emptyTableCellStyle is the minimum TableCellStyle needed to avoid nil
// dereferences in gdocs.Convert (it accesses TableCellStyle.ColumnSpan).
var emptyTableCellStyle = &docs.TableCellStyle{ColumnSpan: 1}

// textCell returns a single-paragraph table cell containing plain text.
func textCell(text string) *docs.TableCell {
	return &docs.TableCell{
		TableCellStyle: emptyTableCellStyle,
		Content: []*docs.StructuralElement{
			{
				Paragraph: &docs.Paragraph{
					ParagraphStyle: normalParagraphStyle,
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: text + "\n"}},
					},
				},
			},
		},
	}
}

// imageCell returns a table cell containing a single inline object reference.
func imageCell(objID string) *docs.TableCell {
	return &docs.TableCell{
		TableCellStyle: emptyTableCellStyle,
		Content: []*docs.StructuralElement{
			{
				Paragraph: &docs.Paragraph{
					ParagraphStyle: normalParagraphStyle,
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: objID}},
					},
				},
			},
		},
	}
}

// makeInlineObject returns the InlineObjects map entry for a given object ID
// pointing at imageURL.
func makeInlineObject(objID, imageURL string) docs.InlineObject {
	return docs.InlineObject{
		ObjectId: objID,
		InlineObjectProperties: &docs.InlineObjectProperties{
			EmbeddedObject: &docs.EmbeddedObject{
				ImageProperties: &docs.ImageProperties{
					ContentUri: imageURL,
				},
			},
		},
	}
}

// wideImageCell returns a table cell with ColumnSpan=2 containing an inline
// object. This mirrors the layout GDocs produces when a lede image is inserted
// on its own row: the key ("lede image") occupies one single-column row, and
// the inline image occupies the next row as a spanning cell.
func wideImageCell(objID string) *docs.TableCell {
	wideStyle := &docs.TableCellStyle{ColumnSpan: 2}
	return &docs.TableCell{
		TableCellStyle: wideStyle,
		Content: []*docs.StructuralElement{
			{
				Paragraph: &docs.Paragraph{
					ParagraphStyle: normalParagraphStyle,
					Elements: []*docs.ParagraphElement{
						{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: objID}},
					},
				},
			},
		},
	}
}

// makeMetadataDocWithLedeImage builds a docs.Document whose only table is a
// metadata table containing a "lede image" block that references an inline
// image. The layout matches what GDocs actually produces:
//
//	"lede image" key is in its own single-column row (no value column).
//	The inline image occupies the NEXT row as a colspan=2 cell.
//
// valueOrNext picks up the image from cells.At(i+1, 0), which is how the
// real code works. No path row is present, so processDocExternals must upload
// and set it.
func makeMetadataDocWithLedeImage(imageURL string) docs.Document {
	const objID = "test-lede-img-obj"
	return docs.Document{
		Title: "Test Metadata Lede Image Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					Table: &docs.Table{
						Rows:    5,
						Columns: 2,
						TableRows: []*docs.TableRow{
							// row 0: label (single colspan=2 cell)
							{TableCells: []*docs.TableCell{labelCell("metadata")}},
							// row 1: Hed
							{TableCells: []*docs.TableCell{textCell("hed"), textCell("Test Headline")}},
							// row 2: "lede image" key only (no value column)
							{TableCells: []*docs.TableCell{textCell("lede image")}},
							// row 3: the image in a spanning cell (valueOrNext falls here)
							{TableCells: []*docs.TableCell{wideImageCell(objID)}},
							// row 4: Lede image credit
							{TableCells: []*docs.TableCell{textCell("lede image credit"), textCell("Test Photographer")}},
						},
					},
				},
			},
		},
		InlineObjects: map[string]docs.InlineObject{
			objID: makeInlineObject(objID, imageURL),
		},
	}
}

// labelCell returns a colspan=2 cell used as the first (label) row of a GDocs
// metadata/photo table.
func labelCell(text string) *docs.TableCell {
	return &docs.TableCell{
		TableCellStyle: &docs.TableCellStyle{ColumnSpan: 2},
		Content: []*docs.StructuralElement{
			{
				Paragraph: &docs.Paragraph{
					ParagraphStyle: normalParagraphStyle,
					Elements: []*docs.ParagraphElement{
						{TextRun: &docs.TextRun{Content: text + "\n"}},
					},
				},
			},
		},
	}
}

// makePhotoDocWithInlineImage builds a docs.Document containing a photo table
// with an inline image and no pre-existing path row.
func makePhotoDocWithInlineImage(imageURL string) docs.Document {
	const objID = "test-photo-obj"
	return docs.Document{
		Title: "Test Photo Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					Table: &docs.Table{
						Rows:    3,
						Columns: 2,
						TableRows: []*docs.TableRow{
							// row 0: label | inline image
							{TableCells: []*docs.TableCell{textCell("photo"), imageCell(objID)}},
							// row 1: credit
							{TableCells: []*docs.TableCell{textCell("credit"), textCell("Test Credit")}},
							// row 2: description
							{TableCells: []*docs.TableCell{textCell("description"), textCell("Test Alt")}},
						},
					},
				},
			},
		},
		InlineObjects: map[string]docs.InlineObject{
			objID: makeInlineObject(objID, imageURL),
		},
	}
}

// TestProcessDocExternalsReplacesImagePath verifies that processDocExternals
// (called internally by ProcessGDocsDoc) uploads images referenced in photo
// and metadata tables and writes the resulting CAS path back into the table,
// so that processDocHTML can pick it up and produce correct embeds/metadata.
func TestProcessDocExternalsReplacesImagePath(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)

	const imageURL = "https://images.example.com/test-photo.jpg"

	newSvc := func() almsvc.Services {
		return almsvc.Services{
			DB:         dbhandle,
			Queries:    dbhandle.Queries(),
			ImageStore: aws.NewBlobStore("mem://"),
			FileStore:  aws.NewBlobStore("mem://"),
			Gsvc:       new(google.Service),
			Client: &http.Client{
				Transport: reqtest.ReplayString(minimalJPEGResponse),
			},
		}
	}

	t.Run("photo table", func(t *testing.T) {
		be := assert.FailsNow(t)
		svc := newSvc()
		svc.Gsvc.SetMockClient(svc.Client)
		ctx := t.Context()

		doc := makePhotoDocWithInlineImage(imageURL)
		dbDoc := be.OK(svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
			ExternalID: "test-img-replace-photo",
			Document:   doc,
		}))
		be.NilError(svc.ProcessGDocsDoc(ctx, dbDoc))
		dbDoc = be.OK(svc.Queries.GetGDocsByID(ctx, dbDoc.ID))

		// The photo embed should be present and have a CAS path.
		be.Equal(len(dbDoc.Embeds), 1)
		imgEmbed, ok := dbDoc.Embeds[0].Value.(db.EmbedImage)
		be.Truthy(ok)
		be.Truthy(imgEmbed.Path)
		be.Truthy(strings.HasPrefix(imgEmbed.Path, "cas/"))
	})

	t.Run("metadata lede image", func(t *testing.T) {
		// Exercises replaceMetadataImagePath, which handles lede images
		// declared in the metadata table via a "lede image" key row followed
		// by a colspan=2 image row (matching actual GDocs document structure).
		be := assert.FailsNow(t)
		svc := newSvc()
		svc.Gsvc.SetMockClient(svc.Client)
		ctx := t.Context()

		doc := makeMetadataDocWithLedeImage(imageURL)
		dbDoc := be.OK(svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
			ExternalID: "test-img-replace-metadata",
			Document:   doc,
		}))
		be.NilError(svc.ProcessGDocsDoc(ctx, dbDoc))
		dbDoc = be.OK(svc.Queries.GetGDocsByID(ctx, dbDoc.ID))

		// processDocExternals must have set the path in the metadata table so
		// that processDocHTML can read it into Metadata.LedeImage.
		be.Truthy(dbDoc.Metadata.LedeImage)
		be.Truthy(strings.HasPrefix(dbDoc.Metadata.LedeImage, "cas/"))
		be.Equal(dbDoc.Metadata.LedeImageCredit, "Test Photographer")
	})
}
