package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
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
	debugLogger     *log.Logger
	infoLogger      *log.Logger
	warnLogger      *log.Logger
	successLogger   *log.Logger
	completedLogger *log.Logger
	errorLogger     *log.Logger
	fatalLogger     *log.Logger
	level           LogLevel
}

func New(level LogLevel) *Logger {
	return NewWithWriter(os.Stdout, level)
}

func NewWithWriter(writer io.Writer, level LogLevel) *Logger {
	return &Logger{
		debugLogger:     log.New(writer, "", 0),
		infoLogger:      log.New(writer, "", 0),
		successLogger:   log.New(writer, "", 0),
		completedLogger: log.New(writer, "", 0),
		warnLogger:      log.New(writer, "", 0),
		errorLogger:     log.New(writer, "", 0),
		fatalLogger:     log.New(writer, "", 0),
		level:           level,
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Output(2, fmt.Sprintf("%s[DEBUG]%s %s", colorGray, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Output(2, fmt.Sprintf("%s[INFO]%s %s", colorBlue, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Success(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.successLogger.Output(2, fmt.Sprintf("%s[SUCCESS]%s %s", colorGreen, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Completed(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.completedLogger.Output(2, fmt.Sprintf("%s[COMPLETED]%s %s", colorGreen, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.warnLogger.Output(2, fmt.Sprintf("%s[WARN]%s %s", colorYellow, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Output(2, fmt.Sprintf("%s[ERROR]%s %s", colorRed, colorReset, fmt.Sprintf(format, v...)))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Output(2, fmt.Sprintf("%s[FATAL]%s %s", colorPurple, colorReset, fmt.Sprintf(format, v...)))
		os.Exit(1)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Default logger instance
var defaultLogger = New(InfoLevel)

func SetDefaultLogLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func Debug(format string, v ...interface{})     { defaultLogger.Debug(format, v...) }
func Info(format string, v ...interface{})      { defaultLogger.Info(format, v...) }
func Success(format string, v ...interface{})   { defaultLogger.Success(format, v...) }
func Completed(format string, v ...interface{}) { defaultLogger.Completed(format, v...) }
func Warn(format string, v ...interface{})      { defaultLogger.Warn(format, v...) }
func Error(format string, v ...interface{})     { defaultLogger.Error(format, v...) }
func Fatal(format string, v ...interface{})     { defaultLogger.Fatal(format, v...) }

type PingRow struct {
	Source      string
	Destination string
	Sent        int
	Received    int
	Loss        float64
	AvgRTT      time.Duration
	Result      string
}

var pingResults []PingRow
var tableHeaderPrinted bool = false

func ClearPingResults() {
	pingResults = []PingRow{}
	tableHeaderPrinted = false
}

func PrintTableHeader() {
	if !tableHeaderPrinted {
		fmt.Println(strings.Repeat("─", 100))
		fmt.Printf("%-18s %-18s %-12s %-14s %-8s  %-12s %-10s\n",
			"Source", "Destination", "Sent", "Received", "Loss(%)", "Avg RTT", "Result")
		fmt.Println(strings.Repeat("─", 100))


		tableHeaderPrinted = true
	}
}

func PingTable(source, dest string, sent, recv int, loss float64, avg time.Duration) {
	result := "Success"
	if recv == 0 {
		result = "Failed"
	} else if recv < sent {
		result = "Partial"
	}

	PrintTableHeader()

	resultColor := colorGreen
	if result == "Failed" {
		resultColor = colorRed
	} else if result == "Partial" {
		resultColor = colorYellow
	}

	lossColor := colorGreen
	if loss == 100.0 {
		lossColor = colorRed
	} else if loss > 0.0 {
		lossColor = colorYellow
	}

	lossStr := fmt.Sprintf("%.1f%%", loss)

	fmt.Printf("%s%-18s%s %-18s %-12d %-14d %s%-8s%s  %-12v %s%-10s%s\n",
		colorBlue, source, colorReset,
		dest, sent, recv,
		lossColor, lossStr, colorReset, 
		avg,
		resultColor, result, colorReset)

	pingResults = append(pingResults, PingRow{
		Source:      source,
		Destination: dest,
		Sent:        sent,
		Received:    recv,
		Loss:        loss,
		AvgRTT:      avg,
		Result:      result,
	})
}

func DisplayPingTable() {
	if !tableHeaderPrinted {
		if len(pingResults) == 0 {
			defaultLogger.Warn("Ping results table is empty")
			return
		}

		PrintTableHeader()

		for _, row := range pingResults {
			resultColor := colorGreen
			if row.Result == "Failed" {
				resultColor = colorRed
			} else if row.Result == "Partial" {
				resultColor = colorYellow
			}

			lossColor := colorGreen
			if row.Loss == 100.0 {
				lossColor = colorRed
			} else if row.Loss > 0.0 {
				lossColor = colorYellow
			}

			fmt.Printf("%s%-18s%s %-18s %-12d %-14d %s%-8.1f%%%s  %-12v %s%-10s%s\n",
				colorBlue, row.Source, colorReset,
				row.Destination, row.Sent, row.Received,
				lossColor, row.Loss, colorReset, 
				row.AvgRTT,
				resultColor, row.Result, colorReset)
		}
	}

	fmt.Println(strings.Repeat("─", 100))
}
