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
	SuccessLevel
	CompletedLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	successLogger *log.Logger
	completedLogger *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       LogLevel
}

func New(level LogLevel) *Logger {
	return NewWithWriter(os.Stdout, level)
}

func NewWithWriter(writer io.Writer, level LogLevel) *Logger {
	return &Logger{
		debugLogger: log.New(writer, "", 0),
		infoLogger:  log.New(writer, "", 0),
		successLogger: log.New(writer, "", 0),
		completedLogger: log.New(writer, "", 0),
		warnLogger:  log.New(writer, "", 0),
		errorLogger: log.New(writer, "", 0),
		fatalLogger: log.New(writer, "", 0),
		level:       level,
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		now := log.Ldate | log.Ltime
		l.debugLogger.Output(2, fmt.Sprintf("%s%s%s %s[DEBUG]%s %s", colorYellow, log.New(os.Stdout, "", now).Prefix(), colorReset, colorGray, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		now := log.Ldate | log.Ltime
		l.infoLogger.Output(2, fmt.Sprintf("%s%s%s %s[INFO]%s %s", colorBlue, log.New(os.Stdout, "", now).Prefix(), colorReset, colorBlue, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Success(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		now := log.Ldate | log.Ltime
		l.infoLogger.Output(2, fmt.Sprintf("%s%s%s %s[SUCCESS]%s %s", colorGreen, log.New(os.Stdout, "", now).Prefix(), colorReset, colorGreen, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Completed(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		now := log.Ldate | log.Ltime
		l.infoLogger.Output(2, fmt.Sprintf("%s%s%s %s[COMPLETED]%s %s", colorGreen, log.New(os.Stdout, "", now).Prefix(), colorReset, colorGreen, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		now := log.Ldate | log.Ltime
		l.warnLogger.Output(2, fmt.Sprintf("%s%s%s %s[WARN]%s %s", colorYellow, log.New(os.Stdout, "", now).Prefix(), colorReset, colorYellow, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		now := log.Ldate | log.Ltime
		l.errorLogger.Output(2, fmt.Sprintf("%s%s%s %s[ERROR]%s %s", colorRed, log.New(os.Stdout, "", now).Prefix(), colorReset, colorRed, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		now := log.Ldate | log.Ltime
		l.fatalLogger.Output(2, fmt.Sprintf("%s%s%s %s[FATAL]%s %s", colorYellow, log.New(os.Stdout, "", now).Prefix(), colorReset, colorPurple, colorReset, fmt.Sprintf(format, v...)))
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

func Success(format string, v ...interface{}) {
	defaultLogger.Success(format, v...)
}

func Completed(format string, v ...interface{}) {
	defaultLogger.Completed(format, v...)
}

func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
