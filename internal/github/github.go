package github

import (
	"context"
	"flag"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

func FlagVar(fl *flag.FlagSet) func(l Logger) (*Client, error) {
	if fl == nil {
		fl = flag.CommandLine
	}

	token := fl.String("github-token", "", "personal access `token` for Github")
	owner := fl.String("github-owner", "", "owning `organization` for Github repo")
	repo := fl.String("github-repo", "", "name of Github `repo`")
	branch := fl.String("github-branch", "", "Github `branch` to use")

	return func(l Logger) (*Client, error) {
		if *token == "" || *owner == "" || *repo == "" || *branch == "" {
			return nil, nil
		}
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc := oauth2.NewClient(ctx, ts)

		client := github.NewClient(tc)
		cl := &Client{client, *owner, *repo, *branch, l}
		if err := cl.Ping(ctx); err != nil {
			return nil, err
		}
		return cl, nil
	}
}

type Client struct {
	client              *github.Client
	owner, repo, branch string
	l                   Logger
}

func (cl *Client) printf(format string, v ...interface{}) {
	if cl.l != nil {
		cl.l.Printf(format, v...)
	}
}

func (cl *Client) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	cl.printf("creating file %s on Github %s/%s@%s",
		path, cl.owner, cl.repo, cl.branch)

	// Note: the file needs to be absent from the repository as you are not
	// specifying a SHA reference here.
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(msg),
		Content: content,
		Branch:  github.String(cl.branch),
	}
	_, _, err := cl.client.Repositories.CreateFile(ctx, cl.owner, cl.repo, path, opts)
	return err
}

func (cl *Client) Ping(ctx context.Context) error {
	cl.printf("pinging Github %s/%s@%s", cl.owner, cl.repo, cl.branch)
	_, _, err := cl.client.Repositories.GetBranch(ctx, cl.owner, cl.repo, cl.branch)
	return err
}
