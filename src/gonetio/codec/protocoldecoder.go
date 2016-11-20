// File ProtocolDecoder
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package codec

import (
	"gonetio"
)

type Decoder interface {
	Decode(*gonetio.IoFilter, gonetio.BaseObject) gonetio.BaseObject
}

type ProtocolDecoder struct {
	gonetio.IoHandlerAdaptor
	decoder Decoder // the real decoder
}

func (this *ProtocolDecoder) SetDecoder(dcoder Decoder) {
	this.decoder = dcoder
}

func (this *ProtocolDecoder) Decode(filter *gonetio.IoFilter, obj gonetio.BaseObject) gonetio.BaseObject {
	return obj
}

func (this *ProtocolDecoder) MessageReceived(filter *gonetio.IoFilter, obj gonetio.BaseObject) {
	for {
		outObject := this.decoder.Decode(filter, obj)
		if outObject != nil {
			filter.MessageReceived(outObject)
		} else {
			break
		}
	}
}
