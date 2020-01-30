module github.com/spotlightpa/almanack

// +heroku goVersion go1.13
// +heroku install ./...

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/aws/aws-lambda-go v1.13.3
	github.com/carlmjohnson/errutil v0.0.9
	github.com/carlmjohnson/exitcode v0.0.3
	github.com/carlmjohnson/flagext v0.0.6
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-redsync/redsync v1.3.1
	github.com/golang/gddo v0.0.0-20200127195332-7365cb292b8b
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	github.com/peterbourgon/ff v1.6.0
	github.com/piotrkubisa/apigo v1.1.2-0.20190907190536-6de39ca9cf97
	github.com/pkg/errors v0.9.1 // indirect
)
