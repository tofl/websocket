package websocket

import (
	"crypto/rand"
)

type Frame struct {
	IsFin         bool
	Opcode        byte
	IsMasked      bool
	PayloadLength uint64
	MaskingKey    [4]byte
	Payload       []byte
}

func NewFrame(isFin bool, opcode byte, isMasked bool, payload []byte) Frame {
	f := Frame{}
	f.IsFin = isFin
	f.Opcode = opcode
	f.IsMasked = isMasked
	f.PayloadLength = uint64(len(payload))

	if f.IsMasked {
		// Generate 4 random bytes
		key := make([]byte, 4)
		rand.Read(key)
		f.MaskingKey = [4]byte(key)
		maskUnmask(&payload, key)
	}

	f.Payload = payload

	return f
}
