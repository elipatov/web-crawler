// Package logger provides logger implementation.
package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Logger is a structured logger.
type Logger struct {
	*slog.Logger
}

// New creates new logger instance.
func New(level string) *Logger {
	var (
		sLevel slog.Level
		levels = map[string]slog.Level{
			"DEBUG": slog.LevelDebug,
			"WARN":  slog.LevelWarn,
			"INFO":  slog.LevelInfo,
			"ERROR": slog.LevelError,
		}
	)

	if l, ok := levels[level]; ok {
		sLevel = l
	} else {
		sLevel = slog.LevelInfo
	}

	return &Logger{
		Logger: slog.New(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				Level: sLevel,
			})),
	}
}

// WithField adds a single field to the Entry.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		l.Logger.With(slog.Attr{
			Key:   key,
			Value: slog.AnyValue(value),
		}),
	}
}

// WithError adds error related fields to the Entry.
func (l *Logger) WithError(err error) *Logger {
	l = l.WithField("error", err)

	if e, ok := err.(interface{ Code() string }); ok {
		l = l.WithField("error_code", e.Code())
	}

	if e, ok := err.(interface{ StackTrace() string }); ok {
		l = l.WithField("stack_trace", e.StackTrace())
	}

	return l
}

func (l *Logger) Infof(msg string, args ...any) {
	l.Info(fmt.Sprintf(msg, args...))
}

// Fatal logs at error level and exits the process with code 1.
func (l *Logger) Fatal(msg string) {
	l.Error(msg)
	os.Exit(1)
}
