// File TcpAcceptor
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

import (
	"net"
	"strconv"
	"sync"
)

type AcceptorConf struct {
	listenPort            int // listen port
	connSendChanSizeLimit int // each connection packet send queue size
	keepAliveMinTime      int // in seconds, the min time duration between two package, valid only when the value is positive
}

// new config
func NewConfig(port int, sendQueueSize int, keepAliveMinTimeDuration int) *AcceptorConf {
	return &AcceptorConf{
		listenPort:            port,
		connSendChanSizeLimit: sendQueueSize,
		keepAliveMinTime:      keepAliveMinTimeDuration,
	}
}

type TcpAcceptor struct {
	config      *AcceptorConf    // the acceptor config
	filterChain *IoFilterChain   // filter chain
	listener    *net.TCPListener // listener
	exitChan    chan struct{}    // notify all goroutines to shutdown
	waitGroup   *sync.WaitGroup  // wait for all goroutines to stop
}

// create new acceptor instance
func NewAcceptor(conf *AcceptorConf) *TcpAcceptor {
	return &TcpAcceptor{
		config:      conf,
		filterChain: NewIoFilterChain(nil),
		listener:    nil,
		exitChan:    make(chan struct{}),
		waitGroup:   &sync.WaitGroup{},
	}
}

// get io filter chain
func (this *TcpAcceptor) GetFilterChain() *IoFilterChain {
	return this.filterChain
}

// the acceptor main loop
func (this *TcpAcceptor) acceptLoop() {

	defer func() {
		if this.listener != nil {
			this.listener.Close()
		}
		this.waitGroup.Done()

		LogInfo("acceptor loop exit")
	}()

	LogInfo("acceptor loop start")

	for {
		select {
		case <-this.exitChan:
			LogInfo("accept loop receive exit signal, exit")
			return
		default:
		}

		conn, err := this.listener.AcceptTCP()
		if err != nil {
			LogError("accept loop, error:%s.", err.Error())
			continue
		}

		addr := conn.RemoteAddr().String()
		LogInfo("accept a new connection[%s].", addr)

		tcpCon := newConn(conn, this)
		tcpCon.SetIoFilterChain(this.filterChain.NewInstanceAndClone(tcpCon))
		tcpCon.Start()
	}

}

// the acceptor start
func (this *TcpAcceptor) Start() bool {

	var err error = nil
	var addr *net.TCPAddr = nil

	addr, err = net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(this.config.listenPort))

	if err != nil {
		LogError("resolve tcp addr failed, port:%d error:%s.", this.config.listenPort, err.Error())
		return false
	}

	this.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		LogError("bind to port[%d] failed, error:%s.", this.config.listenPort, err.Error())
		return false
	}

	LogInfo("acceptor listen to port[%d].", this.config.listenPort)

	this.waitGroup.Add(1)
	go this.acceptLoop()

	return true
}

// stop
func (this *TcpAcceptor) Stop() {
	close(this.exitChan)
}

// wait for stop

func (this *TcpAcceptor) WaitForStop() {
	this.waitGroup.Wait()
}
