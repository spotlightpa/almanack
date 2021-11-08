package clis

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/versioninfo"
	"github.com/spotlightpa/nkotb/pkg/blocko"
)

const NKOTBApp = "NKOTB"

func NKOTB(args []string) error {
	var app nkotbAppEnv
	err := app.ParseArgs(args)
	if err != nil {
		return err
	}
	if err = app.Exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *nkotbAppEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(NKOTBApp, flag.ContinueOnError)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `NKOTB %s - extract blocks of Markdownish content from HTML

Usage:

	nkotb [options] <src>

If not set, src is stdin.

Options:
`, versioninfo.Version)
		fl.PrintDefaults()
		fmt.Fprintln(fl.Output())
	}
	src := flagext.FileOrURL(flagext.StdIO, nil)
	app.src = src
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fl, NKOTBApp); err != nil {
		return err
	}
	if fl.NArg() > 0 {
		if err := src.Set(fl.Arg(0)); err != nil {
			return err
		}
	}
	return nil
}

type nkotbAppEnv struct {
	src io.ReadCloser
}

func (app *nkotbAppEnv) Exec() (err error) {
	defer app.src.Close()
	buf := bufio.NewReader(app.src)
	return blocko.HTMLToMarkdown(os.Stdout, buf)
}
