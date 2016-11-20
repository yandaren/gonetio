// File TcpAcceptor
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

import (
	"fmt"
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

		fmt.Printf("acceptor loop exit\n")
	}()

	fmt.Printf("acceptor loop start\n")

	for {
		select {
		case <-this.exitChan:
			fmt.Printf("accept loop receive exit signal, exit\n")
			return
		default:
		}

		conn, err := this.listener.AcceptTCP()
		if err != nil {
			fmt.Printf("accept loop, error:%s\n", err.Error())
			continue
		}

		addr := conn.RemoteAddr().String()
		fmt.Printf("accept a new connection[%s]\n", addr)

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
		fmt.Printf("resolve tcp addr failed, port:%d error:%s\n", this.config.listenPort, err.Error())
		return false
	}

	this.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		fmt.Printf("bind to port[%d] failed, error:%s\n", this.config.listenPort, err.Error())
		return false
	}

	fmt.Printf("acceptor listen to port[%d]\n", this.config.listenPort)

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
