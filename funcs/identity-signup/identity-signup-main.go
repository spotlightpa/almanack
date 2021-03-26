package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"

	"github.com/carlmjohnson/flagext"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func main() {
	if err := globalEnv.parseEnv(); err != nil {
		globalEnv.sc.Post(slack.Message{
			Attachments: []slack.Attachment{
				{
					Title: "Could not start identity-signup",
					Text:  err.Error(),
					Color: colorRed,
				}}})
		panic(err)
	}
	globalEnv.logger.Printf("starting identity-signup rev %s", almanack.BuildVersion)
	lambda.Start(whitelistEmails)
}

type appEnv struct {
	db     db.Querier
	sc     slack.Client
	logger *log.Logger
}

func (app *appEnv) parseEnv() error {
	app.logger = log.New(os.Stdout, "identity-signup ", log.LstdFlags)
	fl := flag.NewFlagSet("identity-signup", flag.ContinueOnError)
	slackHookURL := fl.String("slack-hook-url", "", "Slack hook endpoint `URL`")
	pg := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	heroku := herokuapi.ConfigureFlagSet(fl)
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	if err := fl.Parse([]string{}); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fl, "almanack"); err != nil {
		return err
	}
	if err := heroku.Configure(app.logger, map[string]string{
		"postgres": "DATABASE_URL",
	}); err != nil {
		return err
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:       *sentryDSN,
		Release:   almanack.BuildVersion,
		Transport: &sentry.HTTPSyncTransport{Timeout: 1 * time.Second},
	}); err != nil {
		return err
	}

	app.sc = slack.New(*slackHookURL, app.logger)
	if err := flagext.MustHave(fl, "postgres"); err != nil {
		return err
	}

	app.db = *pg
	return nil
}

var globalEnv appEnv

const (
	colorGreen = "#78bc20"
	colorRed   = "#da291c"
)

func whitelistEmails(ctx context.Context, request events.APIGatewayProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	defer func() {
		if err != nil {
			sentry.CaptureException(err)
		}
	}()

	var data struct {
		EventType string         `json:"event"`
		User      netlifyid.User `json:"user"`
	}

	if err = json.Unmarshal([]byte(request.Body), &data); err != nil {
		return resp, err
	}

	roles, err := db.GetRolesForEmail(ctx, globalEnv.db, data.User.Email)
	if err != nil {
		return resp, err
	}
	data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles, roles...)

	body, err := json.Marshal(data.User)
	if err != nil {
		return resp, err
	}
	msg := fmt.Sprintf("%s <%s> with %d role(s)",
		data.User.UserMetadata.FullName,
		data.User.Email,
		len(data.User.AppMetadata.Roles))
	color := colorGreen
	if len(data.User.AppMetadata.Roles) < 1 {
		color = colorRed
	}
	globalEnv.sc.Post(
		slack.Message{
			Attachments: []slack.Attachment{
				{
					Title: "New Almanack Registration",
					Text:  msg,
					Color: color,
					Fields: []slack.Field{
						{
							Title: "Roles",
							Value: strings.Join(data.User.AppMetadata.Roles, ", "),
							Short: true,
						}}}}},
	)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
