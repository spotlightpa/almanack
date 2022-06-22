package common

type Logger interface {
	Printf(format string, v ...any)
}
