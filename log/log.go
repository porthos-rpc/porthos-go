package log

import (
	"fmt"
	"time"
)

type LogLevel string

const (
	LInfo    LogLevel = "INFO"
	LDebug   LogLevel = "DEBUG"
	LWarning LogLevel = "WARNING"
	LError   LogLevel = "ERROR"
)

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

func colorForLevel(level LogLevel) string {
	switch level {
	case LError:
		return red
	case LWarning:
		return yellow
	case LInfo:
		return cyan
	case LDebug:
		return magenta
	default:
		return white
	}
}

func doLog(level LogLevel, format string, args ...interface{}) {
	fmt.Printf("%v |%s %s %s| %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		colorForLevel(level), level, reset,
		fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	doLog(LError, format, args...)
}

func Warning(format string, args ...interface{}) {
	doLog(LWarning, format, args...)
}

func Info(format string, args ...interface{}) {
	doLog(LInfo, format, args...)
}

func Debug(format string, args ...interface{}) {
	doLog(LDebug, format, args...)
}
