package db_test

import (
	"encoding/json"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/db"
	"testing"
)

func TestEmbed_UnmarshalJSON(t *testing.T) {
	be := assert.FailNow(t)
	{
		e1 := db.Embed{
			N:    1,
			Type: db.ImageEmbedTag,
			Value: db.EmbedImage{
				Path:        "path",
				Credit:      "credit",
				Caption:     "caption",
				Description: "desc",
			},
		}
		b, err := json.Marshal(e1)
		be.Zero(err)
		var e2 db.Embed
		be.Zero(json.Unmarshal(b, &e2))
		be.Equal(e2, e1)
	}
	{
		e1 := db.Embed{
			N:     2,
			Type:  db.RawEmbedTag,
			Value: "Mork from Ork",
		}
		b, err := json.Marshal(e1)
		be.Zero(err)
		var e2 db.Embed
		be.Zero(json.Unmarshal(b, &e2))
		be.Equal(e2, e1)
	}
	{
		e1 := db.Embed{
			Type: "bad",
		}
		b, err := json.Marshal(e1)
		be.Zero(err)
		var e2 db.Embed
		be.NotZero(json.Unmarshal(b, &e2))
	}
}
