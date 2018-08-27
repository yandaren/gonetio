// File IoFilter
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

import (
	"bytes"
)

type IoFilter struct {
	name    string    // the name of the filter
	handler IoHandler // the handler
	next    *IoFilter // the next filter
	prev    *IoFilter // the pre filter
	conn    *Tcpcon   // the tcp connection
}

// set tcp con
func (flt *IoFilter) SetCon(con *Tcpcon) {
	flt.conn = con
}

// get tcp con
func (flt *IoFilter) GetCon() *Tcpcon {
	return flt.conn
}

// Find next in bound filter
func (flt *IoFilter) findNextInBoundFilter() *IoFilter {
	next := flt.getNext()
	for {
		if next != nil && !next.getHandler().IsInBound() {
			next = next.getNext()
		} else {
			break
		}
	}
	return next
}

// Find next out bound filter
func (flt *IoFilter) findNextOutBoundFilter() *IoFilter {
	prev := flt.getPrev()
	for {
		if prev != nil && !prev.getHandler().IsOutBound() {
			prev = prev.getPrev()
		} else {
			break
		}
	}
	return prev
}

// get the handler of the filter
func (flt *IoFilter) getHandler() IoHandler {
	return flt.handler
}

// set the next filter
func (flt *IoFilter) setNext(filter *IoFilter) {
	flt.next = filter
}

// get the next filter
func (flt *IoFilter) getNext() *IoFilter {
	return flt.next
}

// set the prev filter
func (flt *IoFilter) setPrev(filter *IoFilter) {
	flt.prev = filter
}

// get the previous filter
func (flt *IoFilter) getPrev() *IoFilter {
	return flt.prev
}

// Connection opened
// The event fired when the server accept a new connection
// or the client client server
func (flt *IoFilter) ConnOpened() {
	next := flt.findNextInBoundFilter()
	if next != nil {
		next.getHandler().ConnOpened(next)
	}
}

// Connection closed
func (flt *IoFilter) ConnClosed() {
	next := flt.findNextInBoundFilter()
	if next != nil {
		next.getHandler().ConnClosed(next)
	}
}

// The event fired when receive message from the connection
func (flt *IoFilter) MessageReceived(obj BaseObject) {
	next := flt.findNextInBoundFilter()
	if next != nil {
		next.getHandler().MessageReceived(next, obj)
	}
}

// The event fire write
func (flt *IoFilter) FireWrite(obj BaseObject) {
	next := flt.findNextOutBoundFilter()
	if next != nil {
		next.getHandler().FireWrite(next, obj)
	}
}

// the head filter
type HeadHandler struct {
	IoHandlerImp
}

// new head handler
func newHeadHandler() *HeadHandler {
	handler := &HeadHandler{}
	handler.SetBoundType(OutBound)
	return handler
}

// Fire Write
func (hh *HeadHandler) FireWrite(filter *IoFilter, obj BaseObject) {
	buffer := obj.(*bytes.Buffer)
	filter.GetCon().Flush(buffer, 0)
}

// Clone
func (hh *HeadHandler) Clone() IoHandler {
	return newHeadHandler()
}

// the tail filter
type TailHandler struct {
	IoHandlerImp
}

// new tail handler
func newTailHandler() *TailHandler {
	handler := &TailHandler{}
	handler.SetBoundType(InBound)
	return handler
}

// The event fired when receive message from the connection
func (th *TailHandler) MessageReceived(filter *IoFilter, obj BaseObject) {
	// todo clear the buffer
	buffer := obj.(*bytes.Buffer)
	buffer.Reset()
}

// Clone
func (th *TailHandler) Clone() IoHandler {
	return newTailHandler()
}

type IoFilterChain struct {
	conn *Tcpcon   // the connection
	head *IoFilter // the head io filter
	tail *IoFilter // the tail io filter
}

// New IoFilter Chain Instance
func NewIoFilterChain(con *Tcpcon) *IoFilterChain {

	headFilter := &IoFilter{
		name:    "head",
		handler: newHeadHandler(),
		next:    nil,
		prev:    nil,
		conn:    con,
	}

	tailFilter := &IoFilter{
		name:    "tail",
		handler: newTailHandler(),
		next:    nil,
		prev:    nil,
		conn:    con,
	}

	headFilter.setNext(tailFilter)
	tailFilter.setPrev(headFilter)

	return &IoFilterChain{
		conn: con,
		head: headFilter,
		tail: tailFilter,
	}
}

// Clone
func (fc *IoFilterChain) NewInstanceAndClone(con *Tcpcon) *IoFilterChain {
	chain := NewIoFilterChain(con)

	filter := fc.head.getNext()
	for {
		if filter != fc.tail {
			chain.AddLast(filter.name, filter.getHandler().Clone())
			filter = filter.getNext()
		} else {
			break
		}
	}

	return chain
}

// get head
func (fc *IoFilterChain) GetHeadFilter() *IoFilter {
	return fc.head
}

// get tail
func (fc *IoFilterChain) GetTailFilter() *IoFilter {
	return fc.tail
}

// get the connection
func (fc *IoFilterChain) getCon() *Tcpcon {
	return fc.conn
}

// set the connction
func (fc *IoFilterChain) setCon(con *Tcpcon) {
	filter := fc.head
	for {
		if filter != nil {
			filter.SetCon(con)
			filter = filter.getNext()
		} else {
			break
		}
	}

	fc.conn = con
}

// add last
// add the filter to the last
func (fc *IoFilterChain) AddLast(filterName string, ioHandler IoHandler) {
	filter := &IoFilter{
		name:    filterName,
		handler: ioHandler,
		next:    nil,
		prev:    nil,
		conn:    fc.conn,
	}

	filter.setPrev(fc.tail.getPrev())
	filter.setNext(fc.tail)
	fc.tail.getPrev().setNext(filter)
	fc.tail.setPrev(filter)
}

// fire the connection opened event at the chain
func (fc *IoFilterChain) FireConnOpened() {
	fc.head.ConnOpened()
}

// Connection closed
func (fc *IoFilterChain) FireConnClosed() {
	fc.head.ConnClosed()
}

// The event fired when receive message from the connection
func (fc *IoFilterChain) FireMessageReceived(obj BaseObject) {
	fc.head.MessageReceived(obj)
}

// Fire Write
func (fc *IoFilterChain) FireWrite(obj BaseObject) {
	fc.tail.FireWrite(obj)
}
