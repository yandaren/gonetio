// File log_common.go
// @Author: yandaren1220@126.com
// @Date: 2018-08-27

package gonetio

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Level type
type LogLevel uint32

// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logrus.New()`.
const (
	LvlDebug LogLevel = iota
	LvlInfo
	LvlWarn
	LvlError
	LvlFatal
	LvlNone
)

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level LogLevel) String() string {
	switch level {
	case LvlDebug:
		return "debug"
	case LvlInfo:
		return "info"
	case LvlWarn:
		return "warn"
	case LvlError:
		return "error"
	case LvlFatal:
		return "fatal"
	case LvlNone:
		return "none"
	}

	return "unknown"
}

// ParseLevel takes a string level and returns the Logrus log level constant.
func ParseLevel(lvl string) (LogLevel, error) {
	switch strings.ToLower(lvl) {
	case "none":
		return LvlNone, nil
	case "fatal":
		return LvlFatal, nil
	case "error":
		return LvlError, nil
	case "warn", "warning":
		return LvlWarn, nil
	case "info":
		return LvlInfo, nil
	case "debug":
		return LvlDebug, nil
	}

	var l LogLevel
	return l, fmt.Errorf("not a valid slog Level: %q", lvl)
}

// logger interface
type LoggerHandler interface {
	LogMsg(lvl LogLevel, format string, args ...interface{})
}

type defaultLoggerHandler struct {
	line_sperator string
}

func NewDefaultLoggerHandler() *defaultLoggerHandler {
	logger := &defaultLoggerHandler{}

	if strings.ToLower(runtime.GOOS) == "windows" {
		logger.line_sperator = "\r\n"
	} else {
		logger.line_sperator = "\n"
	}

	return logger
}

func (this *defaultLoggerHandler) LogMsg(lvl LogLevel, format string, args ...interface{}) {
	msg_content := fmt.Sprintf(format, args...)

	// time prefix
	now := time.Now()
	fmt.Printf("[%04d-%02d-%02d %02d:%02d:%02d.%03d][%-5s] %s%s",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000, lvl.String(), msg_content, this.line_sperator)
}

type Logger struct {
	log_lvl     LogLevel      // logger level
	log_handler LoggerHandler // the logger handler
}

func (this *Logger) setLogLvl(lvl LogLevel) {
	this.log_lvl = lvl
}

func (this *Logger) getLogLvl() LogLevel {
	return this.log_lvl
}

func (this *Logger) shouldLog(lvl LogLevel) bool {
	return lvl >= this.log_lvl
}

func (this *Logger) log_msg(lvl LogLevel, format string, args ...interface{}) {
	if !this.shouldLog(lvl) {
		return
	}

	this.log_handler.LogMsg(lvl, format, args...)
}

func (this *Logger) debug(format string, args ...interface{}) {
	this.log_msg(LvlDebug, format, args...)
}

func (this *Logger) info(format string, args ...interface{}) {
	this.log_msg(LvlInfo, format, args...)
}

func (this *Logger) warn(format string, args ...interface{}) {
	this.log_msg(LvlWarn, format, args...)
}

func (this *Logger) error(format string, args ...interface{}) {
	this.log_msg(LvlError, format, args...)
}

func (this *Logger) fatal(format string, args ...interface{}) {
	this.log_msg(LvlFatal, format, args...)
}

var internal_logger *Logger

func init() {
	internal_logger = &Logger{
		log_lvl:     LvlDebug,
		log_handler: NewDefaultLoggerHandler(),
	}
}

func SetLogLvl(lvl LogLevel) {
	internal_logger.setLogLvl(lvl)
}

func SetLogHandler(log_handler LoggerHandler) {
	internal_logger.log_handler = log_handler
}

func LogDebug(format string, args ...interface{}) {
	internal_logger.debug(format, args...)
}

func LogInfo(format string, args ...interface{}) {
	internal_logger.info(format, args...)
}

func LogWarn(format string, args ...interface{}) {
	internal_logger.warn(format, args...)
}

func LogError(format string, args ...interface{}) {
	internal_logger.error(format, args...)
}

func LogFatal(format string, args ...interface{}) {
	internal_logger.fatal(format, args...)
}
