// File IoHandlerAdaptor
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package gonetio

type IoHandlerAdaptor struct {
	IoHandlerImp
}

// Connection opened
// The event fired when the server accept a new connection
// or the client client server
func (this *IoHandlerAdaptor) ConnOpened(filter *IoFilter) {
	filter.ConnOpened()
}

// Connection closed
func (this *IoHandlerAdaptor) ConnClosed(filter *IoFilter) {
	filter.ConnClosed()
}

// The event fired when receive message from the connection
func (this *IoHandlerAdaptor) MessageReceived(filter *IoFilter, obj BaseObject) {
	filter.MessageReceived(obj)
}

// Fire Write
func (this *IoHandlerAdaptor) FireWrite(filter *IoFilter, obj BaseObject) {
	filter.FireWrite(obj)
}
