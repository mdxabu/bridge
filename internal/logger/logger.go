package logger

import (
	"log/slog"
	"os"
)

// Logger is an interface for logging with different log levels
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
}

// SlogAdapter adapts slog.Logger to our Logger interface
type SlogAdapter struct {
	slogger *slog.Logger
}

// NewLogger creates a new Logger that wraps a slog.Logger
func NewLogger(slogger *slog.Logger) Logger {
	return &SlogAdapter{
		slogger: slogger,
	}
}

// Debug logs a debug message
func (l *SlogAdapter) Debug(msg string) {
	l.slogger.Debug(msg)
}

// Debugf logs a formatted debug message
func (l *SlogAdapter) Debugf(format string, args ...interface{}) {
	l.slogger.Debug(format, args...)
}

// Info logs an info message
func (l *SlogAdapter) Info(msg string) {
	l.slogger.Info(msg)
}

// Infof logs a formatted info message
func (l *SlogAdapter) Infof(format string, args ...interface{}) {
	l.slogger.Info(format, args...)
}

// Warn logs a warning message
func (l *SlogAdapter) Warn(msg string) {
	l.slogger.Warn(msg)
}

// Warnf logs a formatted warning message
func (l *SlogAdapter) Warnf(format string, args ...interface{}) {
	l.slogger.Warn(format, args...)
}

// Error logs an error message
func (l *SlogAdapter) Error(msg string) {
	l.slogger.Error(msg)
}

// Errorf logs a formatted error message
func (l *SlogAdapter) Errorf(format string, args ...interface{}) {
	l.slogger.Error(format, args...)
}

// New creates a new logger with the specified level.
func New(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // Default to info
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}

// GetCompatLogger returns a logger that's compatible with the gateway package
func GetCompatLogger(log *slog.Logger) interface{} {
	// Return the logger in the format expected by gateway.New
	// This might be the same logger, or a wrapper around it
	return log
}
