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
	TimeFormat = "15:04:05"
	LogWriter  = os.Stdout
	LogLevel   = INFO
)

func _log(level int, format string, o ...any) {
	if level < LogLevel || LogWriter == nil {
		return
	}
	builder := strings.Builder{}
	if len(TimeFormat) > 0 {
		builder.WriteString(time.Now().Format(TimeFormat) + " ")
	}
	l := fmt.Sprintf(format, o...)
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
