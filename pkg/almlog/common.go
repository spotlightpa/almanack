// Package almlog has the Almanack common logger
package almlog

import (
	"os"
	"time"

	"golang.org/x/exp/slog"
)

var Logger *slog.Logger

var Level = &slog.LevelVar{}

func init() {
	Level.Set(slog.LevelDebug)
}

func removeTime(groups []string, a slog.Attr) slog.Attr {
	// Netlify already logs time
	if a.Key == slog.TimeKey && len(groups) == 0 {
		a.Key = ""
	}
	return a
}

func UseLambdaLogger() {
	opts := slog.HandlerOptions{
		Level:       Level,
		ReplaceAttr: removeTime,
	}
	Logger = slog.New(opts.NewTextHandler(os.Stderr))
	slog.SetDefault(Logger)
}

func shortenTime(groups []string, a slog.Attr) slog.Attr {
	// Omit date from dev
	if a.Key == slog.TimeKey && len(groups) == 0 {
		a.Value = slog.StringValue(a.Value.Time().Format("03:04:05"))
	}
	return a
}

func UseDevLogger() {
	opts := slog.HandlerOptions{
		Level:       Level,
		ReplaceAttr: shortenTime,
	}
	Logger = slog.New(opts.NewTextHandler(colorize{os.Stderr}))
	slog.SetDefault(Logger)
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
