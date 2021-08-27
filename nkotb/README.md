# NKOTB [![GoDoc](https://godoc.org/github.com/spotlightpa/nkotb?status.svg)](https://godoc.org/github.com/spotlightpa/nkotb) [![Go Report Card](https://goreportcard.com/badge/github.com/spotlightpa/nkotb)](https://goreportcard.com/report/github.com/spotlightpa/nkotb)

Extract blocks of Markdownish content from HTML

## Installation

First install [Go](http://golang.org).

If you just want to install the binary to your current directory and don't care about the source code, run

```bash
GOBIN="$(pwd)" go install github.com/spotlightpa/nkotb@latest
```

## Screenshots

```
$ echo '<h1>Hello, <a href="http://example.com">World</a>!</h1><p>This is an <b>example</b>.</p>' | nkotb
# Hello, <a href="http://example.com">World</a>!

This is an <b>example</b>.
```
