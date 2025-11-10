package logging

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}
func Fatal(msg string, args ...any) {
	Logger.Error(msg, args...)
	os.Exit(1)
}
