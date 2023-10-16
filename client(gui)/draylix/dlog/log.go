package dlog

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

var (
	LogStr = []string{
		"TRACE",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"FATAL",
		"OFF",
	}
	LevelNameMap = map[string]int{
		"TRACE": 0,
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  3,
		"ERROR": 4,
		"FATAL": 5,
		"OFF":   6,
	}
)

func String(level int) string {
	if level < 0 || level > len(LogStr) {
		return ""
	}
	return LogStr[level]
}

var (
	TimeFormat = "15:04:05"
	LogWriter  = os.Stdout
	LogLevel   = INFO
)

func LogToFile(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0655)
	if err != nil {
		LogWriter = os.Stdout
	}
	LogWriter = file
}

func _log(level int, format string, o ...any) {
	if level < LogLevel || LogWriter == nil {
		return
	}
	builder := strings.Builder{}
	if len(TimeFormat) > 0 {
		builder.WriteString(time.Now().Format(TimeFormat) + " ")
	}
	l := fmt.Sprintf(format, o...)
	builder.WriteString(LogStr[level] + " ")
	builder.WriteString(l)
	builder.WriteString("\n")
	_, _ = LogWriter.Write([]byte(builder.String()))
}

func Trace(format string, o ...any) {
	_log(TRACE, format, o...)
}
func Debug(format string, o ...any) {
	_log(DEBUG, format, o...)
}
func Info(format string, o ...any) {
	_log(INFO, format, o...)
}
func Warn(format string, o ...any) {
	_log(WARN, format, o...)
}
func Error(format string, o ...any) {
	_log(ERROR, format, o...)
}
func Fatal(format string, o ...any) {
	_log(FATAL, format, o...)
	os.Exit(1)
}
