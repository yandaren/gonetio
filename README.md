# gonetio

a simple tcp netio lib go golang

# examples 
- tcp client
```
package gonetio_test

import (
	"bytes"
	"fmt"
	"gonetio"
	"gonetio/codec"
	"sync"
	"time"
)

type ClientSession struct {
	connector      *gonetio.TcpConnector // connector
	wg             *sync.WaitGroup
	url            string
	reconnectCount int
}

func NewClientSession(name string, remoteUrl string) *ClientSession {
	cnector := gonetio.NewConnector(name, 2048, 0)
	return &ClientSession{
		connector:      cnector,
		wg:             &sync.WaitGroup{},
		url:            remoteUrl,
		reconnectCount: 0,
	}
}

func (this *ClientSession) AsyncConnect() {
	if this.connector != nil {
		this.connector.AsyncConnect(this.url)
	}
}

func (this *ClientSession) TryReconnectAfter(tm time.Duration) {
	go func() {
		timer := time.Tick(tm)
		for {
			_, ok := <-timer
			if ok {
				this.reconnectCount += 1
				fmt.Printf("try reconnect to url[%s], count[%d]\n", this.url, this.reconnectCount)
				this.AsyncConnect()
				return
			} else {
				fmt.Printf("reconnector timer error\n")
			}
		}
	}()
}

func (this *ClientSession) SendData(msg gonetio.BaseObject) {
	this.connector.Write(msg)
}

type SessionEventHandler struct {
	gonetio.IoHandlerImp
	session *ClientSession // the session
}

func newSessionEventHandler(ss *ClientSession) *SessionEventHandler {
	handler := &SessionEventHandler{
		session: ss,
	}
	handler.SetBoundType(gonetio.InBound)
	return handler
}

func (tl *SessionEventHandler) ConnOpened(filter *gonetio.IoFilter) {
	addr := filter.GetCon().RemoteAddr()
	fmt.Printf("connection[%s] opened\n", addr)

	buffer := bytes.NewBufferString("hello world!!!")
	filter.GetCon().Write(buffer)
}

func (tl *SessionEventHandler) ConnClosed(filter *gonetio.IoFilter) {
	addr := filter.GetCon().RemoteAddr()
	fmt.Printf("connection[%s] closed\n", addr)

	tl.session.TryReconnectAfter(5 * time.Second)
}

func (tl *SessionEventHandler) MessageReceived(filter *gonetio.IoFilter, obj gonetio.BaseObject) {

	fmt.Printf("SessionEventHandler -> MessageReceived \n")

	buffer := obj.(*bytes.Buffer)

	data := buffer.Bytes()

	fmt.Printf("recv :%s\n", data)
}

// Clone
func (tl *SessionEventHandler) Clone() gonetio.IoHandler {
	return newSessionEventHandler(nil)
}

func NetioClientTest() {
	fmt.Printf("NetioClientTest\n")

	var url string = "127.0.0.1:8001"
	clientSession := NewClientSession("echoclient", url)
	handler := newSessionEventHandler(clientSession)

	filterChain := clientSession.connector.GetIoFilterChain()
	if filterChain != nil {
		filterChain.AddLast("FrameDecoder", codec.NewFrameDecoder(4, true))
		filterChain.AddLast("FrameEncoder", codec.NewFrameEncoder(4, true))
		filterChain.AddLast("handler", handler)
	} else {
		fmt.Printf("filter chain is nil\n")
	}

	clientSession.AsyncConnect()

	timer := time.Tick(5 * time.Second)
	count := 0
	for {
		_, ok := <-timer
		if ok {
			count += 1
			testData := fmt.Sprintf("data tick count[%d]", count)
			fmt.Printf("send data-> [%s]\n", testData)
			buffer := bytes.NewBufferString(testData)
			clientSession.SendData(buffer)
		} else {
			fmt.Printf("tick timer error\n")
		}
	}
}

```
- tcp server
```
package gonetio_test

import (
	"bytes"
	"fmt"
	"gonetio"
	"gonetio/codec"
)

type EchoEventHandler struct {
	gonetio.IoHandlerImp
}

func newEchoEventHandler() *EchoEventHandler {
	handler := &EchoEventHandler{}
	handler.SetBoundType(gonetio.InBound)
	return handler
}

func (tl *EchoEventHandler) ConnOpened(filter *gonetio.IoFilter) {
	addr := filter.GetCon().RemoteAddr()
	fmt.Printf("connection[%s] opened\n", addr)

	buffer := bytes.NewBufferString("Welcom to this echo server")
	filter.GetCon().Write(buffer)
}

func (tl *EchoEventHandler) ConnClosed(filter *gonetio.IoFilter) {
	addr := filter.GetCon().RemoteAddr()
	fmt.Printf("connection[%s] closed\n", addr)
}

func (tl *EchoEventHandler) MessageReceived(filter *gonetio.IoFilter, obj gonetio.BaseObject) {

	fmt.Printf("EchoEventHandler -> MessageReceived \n")

	buffer := obj.(*bytes.Buffer)

	data := buffer.Bytes()

	fmt.Printf("recv :%s\n", data)

	bufferRet := bytes.NewBuffer(data)
	filter.GetCon().Write(bufferRet)
}

// Clone
func (tl *EchoEventHandler) Clone() gonetio.IoHandler {
	return newEchoEventHandler()
}

func NetioServerTest() {
	fmt.Printf("NetioServerTest\n")

	config := gonetio.NewConfig(8001, 100, 0)

	acceptor := gonetio.NewAcceptor(config)
	handler := newEchoEventHandler()
	acceptor.GetFilterChain().AddLast("FrameDecoder", codec.NewFrameDecoder(4, true))
	acceptor.GetFilterChain().AddLast("FrameEncoder", codec.NewFrameEncoder(4, true))
	acceptor.GetFilterChain().AddLast("handler", handler)

	if !acceptor.Start() {
		fmt.Printf("acceptor start failed\n")
		return
	}

	fmt.Printf("server initialized finished\n")

	acceptor.WaitForStop()
}

```
- customize log handler
```
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

```
