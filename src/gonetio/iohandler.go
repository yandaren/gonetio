// File IoHandler
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

type BaseObject interface {
}

type IoHandler interface {

	// Connection opened
	// The event fired when the server accept a new connection
	// or the client client server
	ConnOpened(*IoFilter)

	// Connection closed
	ConnClosed(*IoFilter)

	// The event fired when receive message from the connection
	MessageReceived(con *IoFilter, obj BaseObject)

	// Fire Write
	FireWrite(con *IoFilter, obj BaseObject)

	// is in bound handler
	IsInBound() bool

	// is out bound handler
	IsOutBound() bool

	// Clone
	Clone() IoHandler
}

var (
	InBound  = 1
	OutBound = 2
)

type IoHandlerImp struct {
	boundType int // in or out bound type
}

// new handler
func newHandler(bType int) *IoHandlerImp {
	return &IoHandlerImp{
		boundType: bType,
	}
}

// Connection opened
// The event fired when the server accept a new connection
// or the client client server
func (this *IoHandlerImp) ConnOpened(*IoFilter) {
}

// Connection closed
func (this *IoHandlerImp) ConnClosed(*IoFilter) {
}

// The event fired when receive message from the connection
func (this *IoHandlerImp) MessageReceived(filter *IoFilter, obj BaseObject) {
}

// Fire Write
func (this *IoHandlerImp) FireWrite(filter *IoFilter, obj BaseObject) {
}

// is in bound handler
func (this *IoHandlerImp) IsInBound() bool {
	return this.boundType&InBound != 0
}

// is out bound handler
func (this *IoHandlerImp) IsOutBound() bool {
	return this.boundType&OutBound != 0
}

// Clone
func (this *IoHandlerImp) Clone() IoHandler {
	return newHandler(InBound)
}

// set bound type
func (this *IoHandlerImp) SetBoundType(bType int) {
	this.boundType = bType
}
