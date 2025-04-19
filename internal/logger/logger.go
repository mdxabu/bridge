package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
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

type Logger struct {
	debugLogger     *log.Logger
	infoLogger      *log.Logger
	successLogger   *log.Logger
	completedLogger *log.Logger
	warnLogger      *log.Logger
	errorLogger     *log.Logger
	fatalLogger     *log.Logger
	level           LogLevel
}

func New(level LogLevel) *Logger {
	return NewWithWriter(os.Stdout, level)
}

func NewWithWriter(writer io.Writer, level LogLevel) *Logger {
	flags := 0
	return &Logger{
		debugLogger:     log.New(writer, "", flags),
		infoLogger:      log.New(writer, "", flags),
		successLogger:   log.New(writer, "", flags),
		completedLogger: log.New(writer, "", flags),
		warnLogger:      log.New(writer, "", flags),
		errorLogger:     log.New(writer, "", flags),
		fatalLogger:     log.New(writer, "", flags),
		level:           level,
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Output(2, color.HiBlackString("[DEBUG] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Output(2, color.CyanString("[INFO] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Success(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.successLogger.Output(2, color.GreenString("[SUCCESS] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Completed(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.completedLogger.Output(2, color.GreenString("[COMPLETED] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.warnLogger.Output(2, color.YellowString("[WARN] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Output(2, color.RedString("[ERROR] ")+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Output(2, color.HiRedString("[FATAL] ")+fmt.Sprintf(format, v...))
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
		fmt.Printf("%-18s %-18s %-12s %-14s %-8s     %-12s %-10s\n",
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

	var resultStyle, lossStyle, sourceStyle *color.Color
	sourceStyle = color.New(color.FgCyan)
	lossStyle = color.New(color.FgGreen)
	resultStyle = color.New(color.FgGreen)

	if result == "Failed" {
		resultStyle = color.New(color.FgRed)
	} else if result == "Partial" {
		resultStyle = color.New(color.FgYellow)
	}

	if loss == 100.0 {
		lossStyle = color.New(color.FgRed)
	} else if loss > 0.0 {
		lossStyle = color.New(color.FgYellow)
	}

	lossStr := fmt.Sprintf("%.1f%%", loss)

	fmt.Printf("%-18s %-18s %-12d %-14d %-8s     %-12s %-10s\n",
		sourceStyle.Sprintf(source),
		dest,
		int(sent),
		int(recv),
		lossStyle.Sprintf("%s", lossStr),
		avg.String(),
		resultStyle.Sprintf(result),
	)

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
	}

	for _, row := range pingResults {
		var resultStyle, lossStyle, sourceStyle *color.Color
		sourceStyle = color.New(color.FgCyan)
		lossStyle = color.New(color.FgGreen)
		resultStyle = color.New(color.FgGreen)

		if row.Result == "Failed" {
			resultStyle = color.New(color.FgRed)
		} else if row.Result == "Partial" {
			resultStyle = color.New(color.FgYellow)
		}

		if row.Loss == 100.0 {
			lossStyle = color.New(color.FgRed)
		} else if row.Loss > 0.0 {
			lossStyle = color.New(color.FgYellow)
		}

		lossStr := fmt.Sprintf("%.1f%%", row.Loss)

		fmt.Printf("%-18s %-18s %-12d %-14d %-8s     %-12s %-10s\n",
			sourceStyle.Sprintf(row.Source),
			row.Destination,
			int(row.Sent),
			int(row.Received),
			lossStyle.Sprintf("%s", lossStr),
			row.AvgRTT.String(),
			resultStyle.Sprintf(row.Result),
		)
	}

	fmt.Println(strings.Repeat("─", 100))
}
