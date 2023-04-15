package testfile

import (
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
	if w := Read(t, wantFile); w == gotStr {
		return
	}
	ext := filepath.Ext(wantFile)
	base := strings.TrimSuffix(wantFile, ext)
	name := base + "-bad" + ext
	Write(t, name, gotStr)
	t.Fatalf("contents of %s != %s", wantFile, name)
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
