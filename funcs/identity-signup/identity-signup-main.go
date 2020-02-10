package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/slack"
)

func main() {
	lambda.Start(whitelistEmails)
}

var (
	whitelisted_domains = os.Getenv("ALMANACK_WHITELIST_DOMAINS")
	slackHookURL        = os.Getenv("ALMANACK_SLACK_HOOK_URL")
	logger              = log.New(os.Stdout, "identity-signup ", log.LstdFlags)
)

func whitelistEmails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var data struct {
		EventType string         `json:"event"`
		User      netlifyid.User `json:"user"`
	}
	err := json.Unmarshal([]byte(request.Body), &data)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	email := strings.ToLower(data.User.Email)

	if strings.HasSuffix(email, "@spotlightpa.org") {
		data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles,
			"Spotlight PA", "arc user")
	}
	if strings.HasSuffix(email, "@inquirer.com") {
		data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles,
			"arc user")
	}
	suffixes := strings.FieldsFunc(whitelisted_domains,
		func(r rune) bool { return r == ',' || r == ' ' })
	for _, suffix := range suffixes {
		if strings.HasSuffix(email, suffix) {
			logger.Printf("%s has domain %s", email, suffix)
			data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles, "editor")
			break
		}
	}

	body, err := json.Marshal(data.User)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	msg := fmt.Sprintf("%s <%s> with %d role(s)",
		data.User.UserMetadata.FullName,
		data.User.Email,
		len(data.User.AppMetadata.Roles))
	slack.New(slackHookURL, logger).Post(
		slack.Message{
			Attachments: []slack.Attachment{
				{
					Title: "New Almanack Registration",
					Text:  msg,
					Color: "#78bc20",
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
