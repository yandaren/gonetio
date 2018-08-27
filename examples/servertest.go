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
