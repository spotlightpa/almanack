module github.com/spotlightpa/almanack

// +heroku goVersion go1.16
// +heroku install ./cmd/...

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/algolia/algoliasearch-client-go/v3 v3.17.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.37.33
	github.com/carlmjohnson/crockford v0.0.3
	github.com/carlmjohnson/emailx v0.20.2
	github.com/carlmjohnson/errutil v0.20.1
	github.com/carlmjohnson/exitcode v0.20.2
	github.com/carlmjohnson/flagext v0.21.0
	github.com/carlmjohnson/gateway v1.20.7
	github.com/carlmjohnson/resperr v0.20.5
	github.com/carlmjohnson/slackhook v0.21.1
	github.com/gabriel-vasile/mimetype v1.2.0
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-chi/chi v1.5.4
	github.com/go-redsync/redsync v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-github/v33 v33.0.0
	github.com/gorilla/feeds v1.1.1
	github.com/lib/pq v1.10.0
	github.com/mattbaird/gochimp v0.0.0-20200820164431-f1082bcdf63f
	gocloud.dev v0.22.0
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4
	golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84
)
