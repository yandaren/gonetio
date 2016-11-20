// File ProtocolEncoder
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package codec

import (
	"gonetio"
)

type Encoder interface {
	Encode(*gonetio.IoFilter, gonetio.BaseObject) gonetio.BaseObject
}

type ProtocolEncoder struct {
	gonetio.IoHandlerAdaptor
	encoder Encoder // real encoder
}

func (this *ProtocolEncoder) SetEncoder(ecoder Encoder) {
	this.encoder = ecoder
}

func (this *ProtocolEncoder) Encode(filter *gonetio.IoFilter, obj gonetio.BaseObject) gonetio.BaseObject {
	return obj
}

func (this *ProtocolEncoder) FireWrite(filter *gonetio.IoFilter, obj gonetio.BaseObject) {
	filter.FireWrite(this.encoder.Encode(filter, obj))
}
