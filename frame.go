package websocket

type Frame struct {
	IsFin         bool
	Opcode        byte
	IsMasked      bool
	PayloadLength uint64
	MaskingKey    [4]byte
	Payload       []byte
}
