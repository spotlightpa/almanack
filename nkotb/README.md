# NKOTB [![GoDoc](https://godoc.org/github.com/spotlightpa/nkotb?status.svg)](https://godoc.org/github.com/spotlightpa/nkotb) [![Go Report Card](https://goreportcard.com/badge/github.com/spotlightpa/nkotb)](https://goreportcard.com/report/github.com/spotlightpa/nkotb)

Turn Google Docs into (new kids on the) blocks of (marky) Markdownish content.

Project consists of three executables: `gdocs`, `nkotb`, and `nkotbweb`.

- `gdocs` is a CLI tool that downloads a Google Doc and converts it to HTML
- `nkotb` is a CLI tool that takes HTML and converts it to Markdownish content
- `nkotbweb` is a web server that runs a web app which converts Google Docs to Markdownish content.

## Credentials

Google has two kinds of credentials. A "Service Account" is like a user with a weird email address. If you share your Document or Drive with that email address, then the service account can access them. Service Account credentials are created at https://console.cloud.google.com/iam-admin/serviceaccounts. The credential consists of a JSON file containing a x509 certificate private key. See https://developers.google.com/accounts/docs/application-default-credentials for more.

The other kind of credential is an OAuth 2.0 Client ID. Using OAuth 2.0 opens a web browser that prompts the user to authorize your account to be able to access the document. Go to https://console.cloud.google.com/apis/credentials and create a new OAuth 2.0 Client ID. The credentials consist of two variables, an oauth client ID and an oauth client secret. See https://developers.google.com/identity/protocols/oauth2 for more.


## Installation

First install [Go](http://golang.org).

If you just want to install the binary to your current directory and don't care about the source code, run

```bash
GOBIN="$(pwd)" go install github.com/spotlightpa/nkotb/...@latest
```

## Screenshots

```
$ echo '<h1>Hello, <a href="http://example.com">World</a>!</h1><p>This is an <b>example</b>.</p>' | nkotb
# Hello, <a href="http://example.com">World</a>!

This is an <b>example</b>.
```

```
$ gdocs -h
gdocs (devel) - extracts a document from Google Docs

Usage:

        gdocs [options]

Uses Google default credentials if no Oauth credentials are provided. See

https://developers.google.com/accounts/docs/application-default-credentials
https://developers.google.com/identity/protocols/oauth2

Options:
  -id string
        ID for Google Doc
  -oauth-client-id id
        client id for Google OAuth 2.0 authentication
  -oauth-client-secret secret
        client secret for Google OAuth 2.0 authentication
  -read-doc path
        path to read document from instead of Google Docs
  -silent
        don't log debug output
  -write-doc path
        path to write out document
```

```
nkotbweb - (devel)

  -client-id Oauth client ID
        Google Oauth client ID
  -client-secret Oauth client secret
        Google Oauth client secret
  -port int
        specify a port to use http rather than AWS Lambda (default -1)
  -sentry-dsn pseudo-URL
        DSN pseudo-URL for Sentry
  -signing-secret secret
        secret for HMAC cookie signing
```
