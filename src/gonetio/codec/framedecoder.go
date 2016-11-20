// File FrameDecodder
// @Author: yandaren1220@126.com
// @Date: 2016-08-23

package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gonetio"
)

const (
	MaxBufferSize = 5 * 1024 * 1024 // max package size 5m
)

const (
	StateReadLength = 0 // state read packet length
	StateReadBody   = 1 // state read packet body
)

type FrameDecoderState struct {
	state  int // current state
	msgLen int // msg length
}

type FrameDecoder struct {
	ProtocolDecoder
	lengthSize        int                // the buffer size of the length info
	containLengthMode bool               // flag weather the msglength contain the length size
	state             *FrameDecoderState // state of the decoder
}

func NewFrameDecoder(packetLengthSize int, containLength bool) *FrameDecoder {
	handler := &FrameDecoder{
		lengthSize:        packetLengthSize,
		containLengthMode: containLength,
		state: &FrameDecoderState{
			state:  StateReadLength,
			msgLen: 0,
		},
	}
	handler.SetBoundType(gonetio.InBound)
	handler.SetDecoder(handler)
	return handler
}

func (this *FrameDecoder) Decode(filter *gonetio.IoFilter, obj gonetio.BaseObject) gonetio.BaseObject {

	inputBuffer := obj.(*bytes.Buffer)
	inputLen := inputBuffer.Len()

	if this.state.state == StateReadLength {
		if inputLen >= this.lengthSize {
			lengthBuffer := inputBuffer.Next(this.lengthSize)
			this.state.msgLen = int(binary.LittleEndian.Uint32(lengthBuffer))

			if this.containLengthMode {
				this.state.msgLen -= this.lengthSize
			}

			if this.state.msgLen < 0 {
				fmt.Printf("FrameDecoder of con[%s], message body size[%d] is negtive, something wrong, force close the connection\n",
					filter.GetCon().RemoteAddr(), this.state.msgLen)

				filter.GetCon().Close()

				return nil
			}

			if this.state.msgLen >= MaxBufferSize {
				fmt.Printf("FrameDecoder of con[%s], message body size[%d] extend max buffer size[%d], something is wrong, force close the connection\n",
					filter.GetCon().RemoteAddr(), this.state.msgLen, MaxBufferSize)

				filter.GetCon().Close()

				return nil
			}

			// change state
			this.state.state = StateReadBody
		}
	}

	if this.state.state == StateReadBody {
		inputLen := inputBuffer.Len()
		if inputLen >= this.state.msgLen {
			msgBodyBuffer := bytes.NewBuffer(inputBuffer.Next(this.state.msgLen))

			this.state.state = StateReadLength

			return msgBodyBuffer
		}
	}

	return nil
}

// Clone
func (this *FrameDecoder) Clone() gonetio.IoHandler {
	return NewFrameDecoder(this.lengthSize, this.containLengthMode)
}
