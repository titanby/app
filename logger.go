package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

const (
	levelDebug = slog.LevelDebug
	levelInfo  = slog.LevelInfo
	levelWarn  = slog.LevelWarn
	levelError = slog.LevelError
	levelFatal = slog.Level(12)
)

var LevelNames = map[slog.Leveler]string{
	levelFatal: "FATAL",
}

type Logger struct {
	l *slog.Logger
}

func setLogLevel(lev slog.Level) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lev,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := LevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	})))
}

func LogWith(args ...any) *Logger {
	return &Logger{slog.With(args...)}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.l.Debug(msg, args...)
}

func LogDebug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.l.Info(msg, args...)
}

func LogInfo(msg string, args ...any) {
	slog.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.l.Warn(msg, args...)
}

func LogWarning(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.l.Error(msg, args...)
}

func LogError(msg string, args ...any) {
	slog.Error(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.l.Log(context.Background(), levelFatal, msg, args...)
	os.Exit(1)
}

func LogFatal(msg string, args ...any) {
	slog.Log(context.Background(), levelFatal, msg, args...)
	os.Exit(1)
}

func (l *Logger) Print(v ...interface{}) {
	if len(v) == 1 {
		l.l.Info(fmt.Sprintf("%v", v[0]))
	} else {
		l.l.With(v[1:]...).Info(fmt.Sprintf("%v", v[0]))
	}
}

func GetLogger() *Logger {
	return &Logger{slog.Default()}
}
