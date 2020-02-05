package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/spotlightpa/almanack/internal/netlifyid"
)

func main() {
	lambda.Start(whitelistEmails)
}

func whitelistEmails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var data struct {
		EventType string         `json:"event"`
		User      netlifyid.User `json:"user"`
	}
	err := json.Unmarshal([]byte(request.Body), &data)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	suffixes := strings.FieldsFunc(os.Getenv("ALMANACK_WHITELIST_DOMAINS"),
		func(r rune) bool { return r == ',' || r == ' ' })
	for _, suffix := range suffixes {
		if strings.HasSuffix(data.User.Email, suffix) {
			fmt.Printf("%s has domain %s", data.User.Email, suffix)
			data.User.AppMetadata.Roles = append(data.User.AppMetadata.Roles, "editor")
			break
		}
	}

	body, err := json.Marshal(data.User)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
