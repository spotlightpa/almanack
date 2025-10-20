package db_test

import (
	"encoding/json"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestEmbed_UnmarshalJSON(t *testing.T) {
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
		be.NilErr(t, err)
		var e2 db.Embed
		be.NilErr(t, json.Unmarshal(b, &e2))
		be.Equal(t, e1, e2)
	}
	{
		e1 := db.Embed{
			N:     2,
			Type:  db.RawEmbedTag,
			Value: "Mork from Ork",
		}
		b, err := json.Marshal(e1)
		be.NilErr(t, err)
		var e2 db.Embed
		be.NilErr(t, json.Unmarshal(b, &e2))
		be.Equal(t, e1, e2)
	}
	{
		e1 := db.Embed{
			Type: "bad",
		}
		b, err := json.Marshal(e1)
		be.NilErr(t, err)
		var e2 db.Embed
		be.Nonzero(t, json.Unmarshal(b, &e2))
	}
}
