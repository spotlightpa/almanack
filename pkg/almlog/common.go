// Package almlog has the Almanack common logger
package almlog

import (
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slog"
)

var IsLambda = atomic.Bool{}

var Level = &slog.LevelVar{}

func replace(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		if IsLambda.Load() {
			// Netlify already logs time
			a.Key = ""
		} else {
			// Omit date from dev
			a.Value = slog.StringValue(a.Value.Time().Format("03:04:05"))
		}
	}
	return a
}

var opts = slog.HandlerOptions{
	Level:       Level,
	ReplaceAttr: replace,
}

var Slogger = slog.New(opts.NewTextHandler(os.Stderr))

func init() {
	slog.SetDefault(Slogger)
	Level.Set(slog.LevelDebug)
}

func LevelThreshold[T time.Duration | int](val, warn, err T) slog.Level {
	if val >= err {
		return slog.LevelError
	}
	if val >= warn {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}
