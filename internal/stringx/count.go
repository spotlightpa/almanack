package stringx

import (
	"github.com/carlmjohnson/bytemap"
)

var whitespace = bytemap.Make(" \t\n\r\v\f")

func WordCount(s string) int {
	n := 0
	inWord := false
	for _, c := range []byte(s) {
		wasInWord := inWord
		inWord = !whitespace[c]
		if !wasInWord && inWord {
			n++
		}
	}
	return n
}

func ColumnInches(s string) float64 {
	return float64(WordCount(s)) / 30
}

func Lines(s string) float64 {
	return float64(WordCount(s)) / 30 * 8
}
