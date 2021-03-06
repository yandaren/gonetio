// File Tcpcon
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

import (
	"bytes"
	"errors"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// error type
var (
	ErrConnShutdown  = errors.New("Connection has shutdown")
	ErrConnClosed    = errors.New("Connection has been closed")
	ErrWriteBlocking = errors.New("Write packet was blocking")
	ErrConnException = errors.New("Connection exception")
)

// connection state
type ConState int32

const (
	ConStateOpened ConState = iota
	ConStateClosed
)

type Tcpcon struct {
	condID           uint32             // connection id
	rawConn          *net.TCPConn       // the raw connection
	keepAliveMinTime int                // in seconds, the min time between two package read from remote, valid only when the value is positive
	customData       interface{}        // save the user custom data
	remoteAddr       string             // the remote addr
	fullBuffer       *bytes.Buffer      // full recv buffer
	recvBuffer       []byte             // recv buffer
	packetSendChan   chan *bytes.Buffer // packet send channel
	conState         ConState           // the connection state
	closeOnce        sync.Once          // make sure the connection call close just once
	closeChan        chan struct{}      // close signal to the read/write loop
	shutdownFlag     int32              // shutdown flag
	ioFilterChain    *IoFilterChain     // filter chain
	waitGroup        *sync.WaitGroup    // wait group
	globalExitChan   chan struct{}      // global exit chan
}

// new a connection instance from tcp acceptor
func newConn(conn *net.TCPConn, aptor *TcpAcceptor) *Tcpcon {
	con := NewConnFull(conn, aptor.config.connSendChanSizeLimit, aptor.waitGroup, aptor.config.keepAliveMinTime)
	con.setGlobalExitChan(aptor.exitChan)
	return con
}

// new a connection instance
func NewConn(conn *net.TCPConn, sendQueueSize int, wg *sync.WaitGroup, keepAliveMinTimeDuration int) *Tcpcon {
	return NewConnFull(conn, sendQueueSize, wg, keepAliveMinTimeDuration)
}

// new a connection instance
func NewConnFull(conn *net.TCPConn, sendQueueSize int, wg *sync.WaitGroup, keepAliveMinTimeDuration int) *Tcpcon {

	var addr string = ""
	if conn != nil {
		addr = conn.RemoteAddr().String()
	}

	return &Tcpcon{
		condID:           0,
		rawConn:          conn,
		keepAliveMinTime: keepAliveMinTimeDuration,
		remoteAddr:       addr,
		fullBuffer:       bytes.NewBuffer([]byte{}),
		recvBuffer:       make([]byte, 65535),
		packetSendChan:   make(chan *bytes.Buffer, sendQueueSize),
		conState:         ConStateClosed,
		shutdownFlag:     0,
		closeChan:        make(chan struct{}),
		ioFilterChain:    nil,
		waitGroup:        wg,
		globalExitChan:   make(chan struct{}),
	}
}

// set conid
func (this *Tcpcon) SetConID(id uint32) {
	this.condID = id
}

// get conid
func (this *Tcpcon) GetConID() uint32 {
	return this.condID
}

// set iofilter chain
func (this *Tcpcon) SetIoFilterChain(chain *IoFilterChain) {
	this.ioFilterChain = chain
	if this.ioFilterChain != nil {
		this.ioFilterChain.setCon(this)
	}
}

// get iofilter chain
func (this *Tcpcon) GetIoFilterChain() *IoFilterChain {
	return this.ioFilterChain
}

// set global exit chan
func (this *Tcpcon) setGlobalExitChan(gexitchan chan struct{}) {
	this.globalExitChan = gexitchan
}

// set custom data
func (this *Tcpcon) SetCustomData(data interface{}) {
	this.customData = data
}

// get custom data
func (this *Tcpcon) GetCustomData() interface{} {
	return this.customData
}

// set remote addr
func (this *Tcpcon) SetRemoteAddr(addr string) {
	this.remoteAddr = addr
}

// get remote addr
func (this *Tcpcon) RemoteAddr() string {
	return this.remoteAddr
}

// is connection opened
func (this *Tcpcon) IsConnected() bool {
	return this.conState == ConStateOpened
}

// is closed
func (this *Tcpcon) IsShutdown() bool {
	return atomic.LoadInt32(&this.shutdownFlag) == 1
}

// shutdown the connection
func (this *Tcpcon) ShutDown() {
	if atomic.SwapInt32(&this.shutdownFlag, 1) == 0 {
		LogError("Tcpcon ShutDown, url[%s].", this.remoteAddr)
		this.Close()
	}
}

// close the connection
func (this *Tcpcon) Close() {
	this.closeOnce.Do(func() {
		this.conState = ConStateClosed
		close(this.closeChan)
		close(this.packetSendChan)
		this.rawConn.Close()

		if !this.IsShutdown() {
			this.ioFilterChain.FireConnClosed()
		}
	})
}

// add to the send queue
func (this *Tcpcon) Flush(buffer *bytes.Buffer, timeout time.Duration) (err error) {
	if this.IsShutdown() {
		return ErrConnShutdown
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrConnException
		}
	}()

	if timeout == 0 {
		select {
		case this.packetSendChan <- buffer:
			return nil

		default:
			return ErrWriteBlocking
		}
	} else {
		select {
		case this.packetSendChan <- buffer:
			return nil
		case <-this.closeChan:
			return ErrConnClosed
		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

// connection start
func (this *Tcpcon) Start() {
	if this.IsShutdown() {
		return
	}

	if this.rawConn != nil {
		this.SetRemoteAddr(this.rawConn.RemoteAddr().String())
	}

	this.conState = ConStateOpened
	this.ioFilterChain.FireConnOpened()

	// start the read/write/handle loop
	asyncDo(this.readLoop, this.waitGroup)
	asyncDo(this.writeLoop, this.waitGroup)
}

// try set read dead line
func (this *Tcpcon) setReadDeadline() {
	if this.keepAliveMinTime > 0 {
		this.rawConn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(this.keepAliveMinTime)))
	}
}

// read
func (this *Tcpcon) read(b []byte) (int, error) {
	return this.rawConn.Read(b)
}

// write
func (this *Tcpcon) Write(obj BaseObject) {
	if this.ioFilterChain != nil {
		this.ioFilterChain.FireWrite(obj)
	} else {
		LogError("Tcpcon Write, but io filter chain is nil, write failed.")
	}
}

// async do
func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}

// read loop
func (this *Tcpcon) readLoop() {
	defer func() {
		if p := recover(); p != nil {
			LogError("panic recover, p: %v", p)
			LogError("stack: %s", debug.Stack())
		}

		this.Close()

		LogInfo("connection[%s] readloop exit.", this.remoteAddr)
	}()

	LogInfo("connection[%s] readloop start.", this.remoteAddr)

	for {
		select {
		case <-this.globalExitChan:
			return
		case <-this.closeChan:
			return

		default:
		}

		this.setReadDeadline()
		readLen, err := this.read(this.recvBuffer)
		if err != nil {
			LogError("connection[%s] read data error, error:%s.", this.remoteAddr, err.Error())
			return
		}

		if readLen == 0 {
			LogError("connection[%s] read data error, read data len is 0, connection may closed.", this.remoteAddr)
			return
		}

		if this.IsShutdown() {
			return
		}

		this.fullBuffer.Write(this.recvBuffer[:readLen])

		if !this.IsShutdown() {
			this.ioFilterChain.FireMessageReceived(this.fullBuffer)
		}
	}
}

// write loop
func (this *Tcpcon) writeLoop() {

	defer func() {
		if p := recover(); p != nil {
			LogError("panic recover, p: %v", p)
			LogError("stack: %s", debug.Stack())
		}

		this.Close()

		LogError("connection[%s] write loop exit.", this.remoteAddr)
	}()

	LogInfo("connection[%s] write loop start.", this.remoteAddr)
	for {
		select {
		case <-this.globalExitChan:
			return
		case <-this.closeChan:
			return
		case p := <-this.packetSendChan:
			if p == nil {
				return
			}
			if this.IsShutdown() {
				return
			}
			if _, err := this.rawConn.Write(p.Bytes()); err != nil {
				return
			}
		}
	}

}
