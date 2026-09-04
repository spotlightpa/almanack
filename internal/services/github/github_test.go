package github_test

import (
	"os"
	"testing"
	"time"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/services/github"
)

func TestGithub(t *testing.T) {
	be := assert.FailNow(t)
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
	be.Zero(client.UpdateFile(ctx, "test create", fname, []byte(testFileContents)))
	// get
	returned := be.OK(client.GetFile(ctx, fname))
	be.Equal(string(returned), testFileContents)
	// update
	testFileContents = time.Now().Format(time.Stamp)
	be.Zero(client.UpdateFile(ctx, "test update", fname, []byte(testFileContents)))
	// get
	returned = be.OK(client.GetFile(ctx, fname))
	be.Equal(string(returned), testFileContents)
}
