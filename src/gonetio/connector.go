// File TcpConnector
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

import (
	"fmt"
	"net"
	"sync"
)

type ConnectorConfig struct {
	sendQueueSize    int            // send queue size
	keepAliveMinTime int            // in seconds, the min time between two package read from remote, valid only when the value is positive
	filterChain      *IoFilterChain // filter chain
}

type TcpConnector struct {
	conn        *Tcpcon          // raw connection
	connName    string           // connection name
	url         string           // connection url
	waitGroup   *sync.WaitGroup  // wait group
	config      *ConnectorConfig // config
	filterChain *IoFilterChain   // filter chain
}

// new a connctor instance
func NewConnector(name string, maxSendQueueSize int, keepAliveTimeDuration int) *TcpConnector {
	wg := &sync.WaitGroup{}
	conf := &ConnectorConfig{
		sendQueueSize:    maxSendQueueSize,
		keepAliveMinTime: keepAliveTimeDuration,
		filterChain:      NewIoFilterChain(nil),
	}

	return &TcpConnector{
		conn:        nil,
		connName:    name,
		url:         "",
		waitGroup:   wg,
		config:      conf,
		filterChain: conf.filterChain,
	}
}

// get tcp
func (this *TcpConnector) GetCon() *Tcpcon {
	return this.conn
}

// write data
func (this *TcpConnector) Write(obj BaseObject) bool {
	if this.conn != nil && !this.conn.IsClosed() {
		this.conn.Write(obj)
		return true
	}
	return false
}

// get iofilter chain
func (this *TcpConnector) GetIoFilterChain() *IoFilterChain {
	return this.filterChain
}

// try connect
func (this *TcpConnector) AsyncConnect(url string) {
	this.waitGroup.Add(1)
	go this.doConnectTask(url)
}

// do connect task
func (this *TcpConnector) doConnectTask(url string) {

	defer func() {
		this.waitGroup.Done()
	}()

	if this.tryConnect(url) {
		this.start()
	} else {
		this.conn.ioFilterChain.FireConnClosed()
	}
}

// try connect to the server
func (this *TcpConnector) tryConnect(url string) bool {
	fmt.Printf("try connect to url[%s]\n", url)

	this.url = url
	this.conn = NewConn(nil, this.config.sendQueueSize, this.waitGroup, this.config.keepAliveMinTime)
	this.conn.SetIoFilterChain(this.config.filterChain)

	addr, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		fmt.Printf("Connection[%s] resolve tcpaddr[%s] failed, error:%s\n", this.connName, url, err.Error())
		return false
	}

	this.conn.rawConn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("Connection[%s] connect to url[%s] failed, error:%s\n", this.connName, url, err.Error())
		return false
	}

	return true
}

// start the connector
func (this *TcpConnector) start() bool {
	if this.conn != nil {
		this.conn.Start()
		return true
	}
	return false
}

// stop the connector
func (this *TcpConnector) Stop() {
	if this.conn != nil {
		this.conn.Close()
	}
}

// is the connector closed
func (this *TcpConnector) IsClosed() bool {
	if this.conn != nil {
		return this.conn.IsClosed()
	}
	return true
}
