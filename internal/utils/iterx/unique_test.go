package iterx_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/iterx"
)

func TestUniqueFunc(t *testing.T) {
	for _, tc := range []struct{ have, want string }{
		{"", ""},
		{"1", "1"},
		{"1", "1"},
		{"1 2", "1 2"},
		{"1 2 2 3", "1 2 3"},
		{"1 2 3 2 3 4", "1 2 3 4"},
	} {
		in := strings.FieldsSeq(tc.have)
		seq := iterx.UniquesFunc(in, func(s string) string { return s })
		got := strings.Join(slices.Collect(seq), " ")
		assert.Continues(t).Equal(got, tc.want)
	}
}
