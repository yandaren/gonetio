// File FrameEncoder
// @Author: fuzuotao@7fgame.com
// @Date: 2016-08-23

package codec

import (
	"bytes"
	"encoding/binary"
	"gonetio"
)

type FrameEncoder struct {
	ProtocolEncoder
	lengthSize        int  // the buffer size of the length info
	containLengthMode bool // flag weather the msglength contain the length size
}

// new FrameEncoder
func NewFrameEncoder(packetLengthSize int, containLength bool) *FrameEncoder {
	handler := &FrameEncoder{
		lengthSize:        packetLengthSize,
		containLengthMode: containLength,
	}
	handler.SetBoundType(gonetio.OutBound)
	handler.SetEncoder(handler)
	return handler
}

func (this *FrameEncoder) Encode(filter *gonetio.IoFilter, obj gonetio.BaseObject) gonetio.BaseObject {
	input := obj.(*bytes.Buffer)

	intputLength := input.Len()
	frameLength := intputLength
	if this.containLengthMode {
		frameLength += this.lengthSize
	}

	frameBuffer := make([]byte, this.lengthSize)
	binary.LittleEndian.PutUint32(frameBuffer[:], uint32(frameLength))

	// write frame buffer
	totalBuffer := bytes.NewBuffer([]byte{})
	totalBuffer.Write(frameBuffer)

	// write body buffer
	totalBuffer.Write(input.Bytes())

	return totalBuffer
}

// Clone
func (this *FrameEncoder) Clone() gonetio.IoHandler {
	return NewFrameEncoder(this.lengthSize, this.containLengthMode)
}
