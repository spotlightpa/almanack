module github.com/spotlightpa/almanack

// +heroku goVersion go1.16
// +heroku install ./cmd/...

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/algolia/algoliasearch-client-go/v3 v3.8.2
	github.com/aws/aws-lambda-go v1.17.0
	github.com/aws/aws-sdk-go v1.36.8
	github.com/carlmjohnson/crockford v0.0.3
	github.com/carlmjohnson/emailx v0.20.2
	github.com/carlmjohnson/errutil v0.20.1
	github.com/carlmjohnson/exitcode v0.0.4
	github.com/carlmjohnson/flagext v0.20.2
	github.com/carlmjohnson/gateway v1.20.6
	github.com/carlmjohnson/resperr v0.20.4
	github.com/carlmjohnson/slackhook v0.20.2
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/getsentry/sentry-go v0.9.0
	github.com/go-chi/chi v1.5.1
	github.com/go-redsync/redsync v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-github/v33 v33.0.0
	github.com/gorilla/feeds v1.1.1
	github.com/lib/pq v1.9.0
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	gocloud.dev v0.21.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/oauth2 v0.0.0-20201203001011-0b49973bad19
)
