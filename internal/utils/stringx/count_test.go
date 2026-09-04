package stringx_test

import (
	_ "embed"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"testing"
)

//go:embed testdata/article.txt
var article string

func TestWordCount(t *testing.T) {
	be := assert.Continue(t)
	cases := []struct {
		s string
		n int
	}{
		{"", 0},
		{" ", 0},
		{"a", 1},
		{"a   'quick' brown fox ", 4},
		{article, 1510},
	}
	for _, tc := range cases {
		be.Equal(stringx.WordCount(tc.s), tc.n)
	}
}

var n int

func BenchmarkWordCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n = stringx.WordCount(article)
	}
}
