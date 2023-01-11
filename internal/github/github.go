// Package github contains utilities for interacting with the Github API.
package github

import (
	"context"
	"errors"
	"flag"
	"net/http"

	"github.com/google/go-github/v48/github"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/stringx"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type ContentStore interface {
	GetFile(ctx context.Context, path string) (content string, err error)
	UpdateFile(ctx context.Context, msg, path string, content []byte) error
}

func AddFlags(fl *flag.FlagSet) func() ContentStore {
	token := fl.String("github-token", "", "personal access `token` for Github")
	owner := fl.String("github-owner", "", "owning `organization` for Github repo")
	repo := fl.String("github-repo", "", "name of Github `repo`")
	branch := fl.String("github-branch", "", "Github `branch` to use")
	mock := fl.String("github-mock-path", "", "`path` for mock Github files")
	return func() ContentStore {
		if *token == "" || *owner == "" || *repo == "" || *branch == "" {
			return NewMockClient(*mock)
		}
		return NewClient(*token, *owner, *repo, *branch)
	}
}

type Client struct {
	client              *github.Client
	owner, repo, branch string
}

func NewClient(token, owner, repo, branch string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	cl := &Client{client, owner, repo, branch}
	/* Omit test ping to keep startup time lower; increase resilience */
	// if err := cl.Ping(ctx); err != nil {
	// 	return nil, err
	// }
	return cl
}

func (cl *Client) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	l := slog.FromContext(ctx)
	l.Info("github.CreateFile",
		"org", cl.owner,
		"repo", cl.repo,
		"branch", cl.branch,
		"path", path,
	)

	// Note: the file needs to be absent from the repository as you are not
	// specifying a SHA reference here.
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(msg),
		Content: content,
		Branch:  github.String(cl.branch),
		Author:  makeAuthor(ctx),
	}
	_, _, err := cl.client.Repositories.CreateFile(ctx, cl.owner, cl.repo, path, opts)
	return err
}

func (cl *Client) GetFile(ctx context.Context, path string) (contents string, err error) {
	l := slog.FromContext(ctx)
	l.Info("github.GetFile",
		"org", cl.owner,
		"repo", cl.repo,
		"branch", cl.branch,
		"path", path,
	)

	fileInfo, _, _, err := cl.client.Repositories.GetContents(
		ctx,
		cl.owner,
		cl.repo,
		path,
		&github.RepositoryContentGetOptions{Ref: cl.branch})
	if err != nil {
		return
	}
	contents, err = fileInfo.GetContent()
	return
}

func (cl *Client) UpdateFile(ctx context.Context, msg, path string, content []byte) error {
	l := slog.FromContext(ctx)
	l.Info("github.UpdateFile",
		"org", cl.owner,
		"repo", cl.repo,
		"branch", cl.branch,
		"path", path,
	)

	fileInfo, _, _, err := cl.client.Repositories.GetContents(
		ctx,
		cl.owner,
		cl.repo,
		path,
		&github.RepositoryContentGetOptions{Ref: cl.branch})
	var sha *string
	if err == nil {
		sha = fileInfo.SHA
		if oldcontent, err2 := fileInfo.GetContent(); err2 == nil && string(content) == oldcontent {
			l.Info("github.UpdateFile skipping; already updated",
				"org", cl.owner,
				"repo", cl.repo,
				"branch", cl.branch,
				"path", path,
			)
			return nil
		}
	} else {
		resp := new(github.ErrorResponse)
		if !errors.As(err, &resp) || resp.Response.StatusCode != http.StatusNotFound {
			return err
		}
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(msg),
		Content: content,
		Branch:  github.String(cl.branch),
		SHA:     sha,
		Author:  makeAuthor(ctx),
	}

	_, _, err = cl.client.Repositories.UpdateFile(ctx, cl.owner, cl.repo, path, opts)

	return err
}

func (cl *Client) Ping(ctx context.Context) error {
	l := slog.FromContext(ctx)
	l.Info("github.Ping",
		"org", cl.owner,
		"repo", cl.repo,
		"branch", cl.branch,
	)
	_, _, err := cl.client.Repositories.GetBranch(ctx, cl.owner, cl.repo, cl.branch, true)
	return err
}

func makeAuthor(ctx context.Context) *github.CommitAuthor {
	jwt := netlifyid.FromContext(ctx)
	name := stringx.First(jwt.Username(), "Almanack")
	email := stringx.First(jwt.Email(), "webmaster@spotlightpa.org")

	return &github.CommitAuthor{
		Name:  github.String(name),
		Email: github.String(email),
	}
}
