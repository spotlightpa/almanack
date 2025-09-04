package github_test

import (
	"os"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/github"
)

func TestGithub(t *testing.T) {
	token := os.Getenv("ALMANACK_GITHUB_TEST_TOKEN")
	owner := os.Getenv("ALMANACK_GITHUB_TEST_OWNER")
	repo := os.Getenv("ALMANACK_GITHUB_TEST_REPO")
	branch := os.Getenv("ALMANACK_GITHUB_TEST_BRANCH")

	if token == "" || owner == "" || repo == "" || branch == "" {
		t.Skip("Missing Github ENV vars")
	}

	client := github.NewClient(token, owner, repo, branch)
	ctx := t.Context()
	// create
	testFileContents := time.Now().Format(time.Stamp)
	fname := time.Now().Format("test-" + time.RFC3339 + ".txt")
	err := client.UpdateFile(ctx, "test create", fname, []byte(testFileContents))
	be.NilErr(t, err)
	// get
	returned, err := client.GetFile(ctx, fname)
	be.NilErr(t, err)
	be.Equal(t, testFileContents, string(returned))
	// update
	testFileContents = time.Now().Format(time.Stamp)
	err = client.UpdateFile(ctx, "test update", fname, []byte(testFileContents))
	be.NilErr(t, err)
	// get
	returned, err = client.GetFile(ctx, fname)
	be.NilErr(t, err)
	be.Equal(t, testFileContents, string(returned))
}
