package gonetio_test

import (
	"fmt"
	"gonetio"
	"runtime"
	"strings"
	"time"
)

type TestLoggerHandler struct {
	line_sperator string
}

func (this *TestLoggerHandler) LogMsg(lvl gonetio.LogLevel, format string, args ...interface{}) {
	msg_content := fmt.Sprintf(format, args...)

	// time prefix
	now := time.Now()
	fmt.Printf("[gonetio][%04d-%02d-%02d %02d:%02d:%02d.%03d][%s] %s%s",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000, lvl.String(), msg_content, this.line_sperator)
}

func NewTestLoggerHandler() *TestLoggerHandler {
	handler := &TestLoggerHandler{}
	if strings.ToLower(runtime.GOOS) == "windows" {
		handler.line_sperator = "\r\n"
	} else {
		handler.line_sperator = "\n"
	}
	return handler
}

func TestUserCustomLoggerHandler() {

	gonetio.SetLogLvl(gonetio.LvlDebug)
	gonetio.SetLogHandler(NewTestLoggerHandler())

	gonetio.LogDebug("gonetio test log debug")
	gonetio.LogInfo("gonetio test log info")
	gonetio.LogWarn("gonetio test log warn")
	gonetio.LogError("gonetio test log error")
	gonetio.LogFatal("gonetio test log fatal")
}
