// Package testfile has test helpers that work by comparing files.
package testfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Read(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return string(b)
}

func Write(t testing.TB, name, data string) {
	t.Helper()
	dir := filepath.Dir(name)
	_ = os.MkdirAll(dir, 0700)
	err := os.WriteFile(name, []byte(data), 0644)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func Equal(t testing.TB, wantFile, gotStr string) {
	t.Helper()
	equal(t, wantFile, gotStr, false)
}

func Equalish(t testing.TB, wantFile, gotStr string) {
	t.Helper()
	equal(t, wantFile, gotStr, true)
}

func equal(t testing.TB, wantFile, gotStr string, trim bool) {
	t.Helper()
	w := Read(t, wantFile)
	if trim {
		w = strings.TrimSpace(w)
		gotStr = strings.TrimSpace(gotStr)
	}
	if w == gotStr {
		return
	}
	ext := filepath.Ext(wantFile)
	base := strings.TrimSuffix(wantFile, ext)
	name := base + "-bad" + ext
	Write(t, name, gotStr)
	t.Fatalf("contents of %s != %s", wantFile, name)
}

func EqualJSON(t testing.TB, wantFile string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshaling: %v", err)
		return
	}
	Equalish(t, wantFile, string(b))
}

func GlobRun(t *testing.T, pat string, f func(path string, t *testing.T)) {
	t.Helper()
	matches, err := filepath.Glob(pat)
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i := range matches {
		path := matches[i]
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			f(path, t)
		})
	}
}

func ReadJSON(t testing.TB, name string, v any) {
	t.Helper()
	s := Read(t, name)
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
}
