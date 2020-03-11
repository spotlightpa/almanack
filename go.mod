module github.com/spotlightpa/almanack

// +heroku goVersion go1.14
// +heroku install ./cmd/...

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go-v2 v0.19.0
	github.com/carlmjohnson/crockford v0.0.3
	github.com/carlmjohnson/errutil v0.0.9
	github.com/carlmjohnson/exitcode v0.0.4
	github.com/carlmjohnson/flagext v0.0.10
	github.com/carlmjohnson/slackhook v0.0.3
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-redsync/redsync v1.3.1
	github.com/golang/gddo v0.0.0-20200127195332-7365cb292b8b
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-github/v29 v29.0.2
	github.com/lib/pq v1.3.0
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	github.com/peterbourgon/ff/v2 v2.0.0
	github.com/piotrkubisa/apigo v1.1.2-0.20190907190536-6de39ca9cf97
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
)
