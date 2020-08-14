module github.com/spotlightpa/almanack

// +heroku goVersion go1.14
// +heroku install ./cmd/...

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/algolia/algoliasearch-client-go/v3 v3.8.2
	github.com/aws/aws-lambda-go v1.17.0
	github.com/aws/aws-sdk-go-v2 v0.19.0
	github.com/carlmjohnson/crockford v0.0.3
	github.com/carlmjohnson/emailx v0.20.2
	github.com/carlmjohnson/errutil v0.20.1
	github.com/carlmjohnson/exitcode v0.0.4
	github.com/carlmjohnson/flagext v0.0.10
	github.com/carlmjohnson/gateway v1.20.6
	github.com/carlmjohnson/resperr v0.20.4
	github.com/carlmjohnson/slackhook v0.20.2
	github.com/getsentry/sentry-go v0.6.1
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-redsync/redsync v1.3.1
	github.com/golang/gddo v0.0.0-20200127195332-7365cb292b8b
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-github/v30 v30.0.0
	github.com/lib/pq v1.3.0
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	github.com/peterbourgon/ff/v2 v2.0.0
	github.com/tj/assert v0.0.3 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
)
