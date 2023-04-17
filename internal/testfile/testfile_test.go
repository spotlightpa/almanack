package testfile_test

import (
	"testing"

	"github.com/spotlightpa/almanack/internal/testfile"
)

func TestJSON(t *testing.T) {
	testfile.GlobRun(t, "testdata/*.json", func(path string, t *testing.T) {
		testfile.EqualJSON(t, path, struct {
			Data any `json:"data"`
		}{
			Data: []struct {
				Field string `json:"field"`
				Value int    `json:"value"`
			}{
				{"foo", 1},
				{"bar", 2},
			},
		})
	})

}
