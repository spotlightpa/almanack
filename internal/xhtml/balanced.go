package xhtml

import (
	"strings"

	"github.com/carlmjohnson/deque"
	"golang.org/x/net/html"
)

// IsBalanced reports whether every opening tag has a closing pair.
func IsBalanced(s string) bool {
	r := strings.NewReader(s)
	z := html.NewTokenizer(r)
	depth := 0
	open := deque.Of[string]()
	for {
		switch tt := z.Next(); tt {
		case html.ErrorToken:
			return depth == 0
		case html.StartTagToken:
			depth++
			tag, _ := z.TagName()
			open.PushBack(string(tag))
		case html.EndTagToken:
			depth--
			tag, _ := z.TagName()
			wantTag, ok := open.RemoveBack()
			if !ok || wantTag != string(tag) {
				return false
			}
		}
	}
}
