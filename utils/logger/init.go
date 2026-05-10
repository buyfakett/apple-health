package logger

import (
	"github.com/gookit/slog"
)

func InitLog(logLevel string) {
	var lvl slog.Level
	switch logLevel {
	case "debug":
		lvl = slog.DebugLevel
	case "info":
		lvl = slog.InfoLevel
	case "warn", "warning":
		lvl = slog.WarnLevel
	case "error":
		lvl = slog.ErrorLevel
	default:
		lvl = slog.InfoLevel
	}

	slog.Configure(func(logger *slog.SugaredLogger) {
		if f, ok := logger.Formatter.(*slog.TextFormatter); ok {
			f.EnableColor = true
			f.TimeFormat = "2006-01-02 15:04:05.000"
		}
		slog.SetLogLevel(lvl)
	})
}
