package almlog

import (
	"io"
	"log/slog"
)

func UseTestLogger(t interface{ Output() io.Writer }) {
	opts := slog.HandlerOptions{
		Level:       Level,
		ReplaceAttr: removeTime,
	}
	Logger = slog.New(slog.NewTextHandler(t.Output(), &opts))
	slog.SetDefault(Logger)
}
