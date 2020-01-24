module github.com/spotlightpa/almanack

// +heroku goVersion go1.13
// +heroku install ./...

go 1.13

require (
	github.com/aws/aws-lambda-go v1.13.3
	github.com/carlmjohnson/errutil v0.0.9
	github.com/carlmjohnson/exitcode v0.0.3
	github.com/carlmjohnson/flagext v0.0.6
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-redsync/redsync v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	github.com/peterbourgon/ff v1.6.0
	github.com/piotrkubisa/apigo v2.0.0+incompatible
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tj/assert v0.0.0-20190920132354-ee03d75cd160 // indirect
	golang.org/x/net v0.0.0-20181201002055-351d144fa1fc // indirect
	golang.org/x/text v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898
)
