package logger

import (
	"log/slog"
	"os"
)

type slogger struct {
	logger *slog.Logger
}

// NewSlog creates a new structured logger
func NewSlog(level string) Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	return &slogger{
		logger: slog.New(handler),
	}
}

func (s *slogger) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s *slogger) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s *slogger) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

func (s *slogger) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s *slogger) With(args ...any) Logger {
	return &slogger{
		logger: s.logger.With(args...),
	}
}
