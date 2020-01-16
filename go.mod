module github.com/spotlightpa/almanack

// +heroku goVersion go1.13
// +heroku install ./...

go 1.13

require (
	github.com/apex/gateway v1.1.1
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go-v2 v0.18.0 // indirect
	github.com/carlmjohnson/errutil v0.0.9
	github.com/carlmjohnson/exitcode v0.0.3
	github.com/carlmjohnson/flagext v0.0.6
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/mattbaird/gochimp v0.0.0-20180111040707-a267553896d1
	github.com/peterbourgon/ff v1.6.0
	github.com/tj/assert v0.0.0-20190920132354-ee03d75cd160 // indirect
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898
)

replace github.com/apex/gateway => github.com/carlmjohnson/gateway v1.1.2-0.20200116185330-eb97f9f4ca3c
