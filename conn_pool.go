// File Tcpcon
// @Author: yandaren1220@126.com
// @Date: 2018-10-16

package gonetio

import (
	"sync"
)

type TcpconnectionPool struct {
	/* <conid, *Tcpcon */
	connection_map map[uint32]*Tcpcon //connection map

	map_mtx *sync.RWMutex // the mutex of the config
}

func NewTcpconnectionPool() *TcpconnectionPool {
	return &TcpconnectionPool{
		connection_map: make(map[uint32]*Tcpcon),
		map_mtx:        &sync.RWMutex{},
	}
}

func (this *TcpconnectionPool) AddCon(con *Tcpcon) uint32 {

	if con == nil {
		return 0
	}

	this.map_mtx.Lock()
	defer this.map_mtx.Unlock()

	this.connection_map[con.GetConID()] = con

	LogDebug("TcpconnectionPool add con[%d] addr[%s], cur pool size[%d]",
		con.GetConID(), con.RemoteAddr(), len(this.connection_map))

	return con.GetConID()
}

func (this *TcpconnectionPool) RemoveCon(con *Tcpcon) {
	if con == nil {
		return
	}

	this.map_mtx.Lock()
	defer this.map_mtx.Unlock()

	delete(this.connection_map, con.GetConID())

	LogDebug("TcpconnectionPool remove con[%d] addr[%s], remain pool size[%d]",
		con.GetConID(), con.RemoteAddr(), len(this.connection_map))
}

func (this *TcpconnectionPool) Send(con_id uint32, obj BaseObject) {
	this.map_mtx.RLock()
	defer this.map_mtx.RUnlock()

	con := this.connection_map[con_id]
	if con == nil {
		LogWarn("TcpconnectionPool can't find con of con_id[%d] try send msg failed.", con_id)
		return
	}

	con.Write(obj)
}

func (this *TcpconnectionPool) Broadcast(obj BaseObject) {
	this.map_mtx.RLock()
	defer this.map_mtx.RUnlock()

	for _, con := range this.connection_map {
		con.Write(obj)
	}
}

func (this *TcpconnectionPool) Size() int32 {
	this.map_mtx.RLock()
	defer this.map_mtx.RUnlock()

	return int32(len(this.connection_map))
}

func (this *TcpconnectionPool) Close(con_id uint32) {
	this.map_mtx.RLock()
	defer this.map_mtx.RUnlock()

	con := this.get_con(con_id)
	if con != nil {
		con.Close()
	}
}

func (this *TcpconnectionPool) RemoteAddr(con_id uint32) string {
	this.map_mtx.RLock()
	defer this.map_mtx.RUnlock()

	con := this.get_con(con_id)
	if con != nil {
		return con.RemoteAddr()
	}
	return ""
}

func (this *TcpconnectionPool) get_con(con_id uint32) *Tcpcon {
	return this.connection_map[con_id]
}
