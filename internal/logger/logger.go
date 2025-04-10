package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	ErrorLevel
	FatalLevel
)

type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       LogLevel
}

func New(level LogLevel) *Logger {
	return NewWithWriter(os.Stdout, level)
}

func NewWithWriter(writer io.Writer, level LogLevel) *Logger {
	return &Logger{
		debugLogger: log.New(writer, "[DEBUG] ", log.Ldate|log.Ltime),
		infoLogger:  log.New(writer, "[INFO] ", log.Ldate|log.Ltime),
		warnLogger:  log.New(writer, "[WARN] ", log.Ldate|log.Ltime),
		errorLogger: log.New(writer, "[ERROR] ", log.Ldate|log.Ltime),
		fatalLogger: log.New(writer, "[FATAL] ", log.Ldate|log.Ltime),
		level:       level,
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.warnLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Output(2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

var defaultLogger = New(InfoLevel)

func SetDefaultLogLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
