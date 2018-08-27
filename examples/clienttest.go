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
