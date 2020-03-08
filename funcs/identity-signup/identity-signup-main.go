package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/peterbourgon/ff/v2"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/slack"
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
	openPG := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	checkHerokuPG := herokuapi.FlagVar(fl, "postgres")
	if err := ff.Parse(fl, []string{}, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return err
	}
	app.sc = slack.New(*slackHookURL, app.logger)
	if usedHeroku, err := checkHerokuPG(); err != nil {
		return err
	} else if usedHeroku {
		app.logger.Printf("used Heroku")
	}
	var err error
	if app.db, err = openPG(); err != nil {
		return err
	}
	return nil
}

var globalEnv appEnv

const (
	colorGreen = "#78bc20"
	colorRed   = "#da291c"
)

func whitelistEmails(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data struct {
		EventType string         `json:"event"`
		User      netlifyid.User `json:"user"`
	}
	err := json.Unmarshal([]byte(request.Body), &data)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	roles, err := db.GetRolesForEmailDomain(ctx, globalEnv.db, data.User.Email)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles, roles...)

	body, err := json.Marshal(data.User)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
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
