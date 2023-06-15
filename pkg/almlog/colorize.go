package almlog

import (
	"bytes"
	"io"

	"github.com/go-logfmt/logfmt"
)

const (
	reset      = "\033[0m"
	bold       = "\033[1m"
	dim        = "\033[2m"
	standout   = "\033[3m"
	underscore = "\033[4m"
	blink      = "\033[5m"
	blinkmore  = "\033[6m"
	invert     = "\033[7m"
	hide       = "\033[8m"
	del        = "\033[9m"
	black      = "\033[30m"
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	blue       = "\033[34m"
	magenta    = "\033[35m"
	cyan       = "\033[36m"
	white      = "\033[37m"
	purple     = magenta + bold
)

type colorize struct {
	io.Writer
}

func (c colorize) Write(p []byte) (int, error) {
	var buf bytes.Buffer
	buf.Grow(len(p))
	d := logfmt.NewDecoder(bytes.NewReader(p))
	for d.ScanRecord() {
		for d.ScanKeyval() {
			skipKey := false
			valColor := purple
			switch string(d.Key()) {
			case "time":
				skipKey = true
				valColor = dim
			case "msg":
				skipKey = true
				valColor = white + underscore
			case "level":
				skipKey = true
				switch string(d.Value()) {
				case "DEBUG":
					valColor = dim
				case "INFO":
					valColor = green + " "
				case "WARN":
					valColor = yellow + " "
				case "ERROR":
					valColor = red
				}
			case "duration":
				valColor = magenta
			case "err":
				valColor = red + bold
			case "res_status":
				valColor = white
			case "req_method":
				valColor = white
			case "req_path":
				valColor = white
			}
			if !skipKey {
				buf.WriteString(cyan)
				buf.Write(d.Key())
				buf.WriteString(yellow)
				buf.WriteString("=")
				buf.WriteString(reset)
			}
			buf.WriteString(valColor)
			buf.Write(d.Value())
			buf.WriteString(reset)
			buf.WriteString(" ")
		}
		buf.WriteString("\n")
	}
	if d.Err() != nil {
		return 0, d.Err()
	}
	return c.Writer.Write(buf.Bytes())
}
