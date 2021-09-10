package github

import (
	"context"
	"errors"
	"flag"
	"net/http"

	"github.com/google/go-github/v39/github"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/common"
	"golang.org/x/oauth2"
)

func FlagVar(fl *flag.FlagSet) func(l common.Logger) (common.ContentStore, error) {
	if fl == nil {
		fl = flag.CommandLine
	}

	token := fl.String("github-token", "", "personal access `token` for Github")
	owner := fl.String("github-owner", "", "owning `organization` for Github repo")
	repo := fl.String("github-repo", "", "name of Github `repo`")
	branch := fl.String("github-branch", "", "Github `branch` to use")
	mock := fl.String("github-mock-path", "", "`path` for mock Github files")
	return func(l common.Logger) (common.ContentStore, error) {
		if *token == "" || *owner == "" || *repo == "" || *branch == "" {
			return NewMockClient(*mock, l)
		}
		return NewClient(*token, *owner, *repo, *branch, l)
	}
}

type Client struct {
	client              *github.Client
	owner, repo, branch string
	l                   common.Logger
}

func NewClient(token, owner, repo, branch string, l common.Logger) (*Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	cl := &Client{client, owner, repo, branch, l}
	if err := cl.Ping(ctx); err != nil {
		return nil, err
	}
	return cl, nil
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
		Author:  makeAuthor(ctx),
	}
	_, _, err := cl.client.Repositories.CreateFile(ctx, cl.owner, cl.repo, path, opts)
	return err
}

func (cl *Client) GetFile(ctx context.Context, path string) (contents string, err error) {
	cl.printf("getting file %s from Github %s/%s@%s",
		path, cl.owner, cl.repo, cl.branch)

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
	cl.printf("updating file %s on Github %s/%s@%s",
		path, cl.owner, cl.repo, cl.branch)

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
			cl.printf("file %s already updated", path)
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
	cl.printf("pinging Github %s/%s@%s", cl.owner, cl.repo, cl.branch)
	_, _, err := cl.client.Repositories.GetBranch(ctx, cl.owner, cl.repo, cl.branch, true)
	return err
}

func makeAuthor(ctx context.Context) *github.CommitAuthor {
	jwt := netlifyid.FromContext(ctx)
	name := jwt.Username()
	if name == "" {
		name = "Almanack"
	}
	email := jwt.Email()
	if email == "" {
		email = "webmaster@spotlightpa.org"
	}
	return &github.CommitAuthor{
		Name:  github.String(name),
		Email: github.String(email),
	}
}
