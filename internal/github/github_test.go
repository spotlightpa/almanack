package github_test

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/spotlightpa/almanack/internal/github"
)

func check(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", msg, err)
	}
}

func eq(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Errorf("want %q; have %q", want, have)
	}
}

func TestGithub(t *testing.T) {
	token := os.Getenv("ALMANACK_GITHUB_TEST_TOKEN")
	owner := os.Getenv("ALMANACK_GITHUB_TEST_OWNER")
	repo := os.Getenv("ALMANACK_GITHUB_TEST_REPO")
	branch := os.Getenv("ALMANACK_GITHUB_TEST_BRANCH")

	if token == "" || owner == "" || repo == "" || branch == "" {
		t.Skip("Missing Github ENV vars")
	}
	var buf bytes.Buffer
	l := log.New(&buf, "", 0)
	client, err := github.NewClient(token, owner, repo, branch, l)
	check(t, err, "could not create client")
	ctx := context.Background()
	// create
	testFileContents := time.Now().Format(time.Stamp)
	fname := time.Now().Format("test-" + time.RFC3339 + ".txt")
	err = client.UpdateFile(ctx, "test create", fname, []byte(testFileContents))
	check(t, err, "could not create file")
	// get
	returned, err := client.GetFile(ctx, fname)
	check(t, err, "could not get file")
	eq(t, testFileContents, string(returned))
	// update
	testFileContents = time.Now().Format(time.Stamp)
	err = client.UpdateFile(ctx, "test update", fname, []byte(testFileContents))
	check(t, err, "could not update file")
	// get
	returned, err = client.GetFile(ctx, fname)
	check(t, err, "could not get file")
	eq(t, testFileContents, string(returned))
}
