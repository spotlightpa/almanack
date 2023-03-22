package stringx_test

import (
	_ "embed"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/stringx"
)

//go:embed testdata/article.txt
var article string

func TestWordCount(t *testing.T) {
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
		be.Equal(be.Relaxed(t), tc.n, stringx.WordCount(tc.s))
	}
}

var n int

func BenchmarkWordCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n = stringx.WordCount(article)
	}
}
