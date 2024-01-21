package websocket

import (
	"encoding/binary"
	"net"
)

type Connection struct {
	Conn       *net.Conn
	Origin     string
	Protocol   string
	Extensions []string
	isOpen     bool
	Id         string
}

func (c *Connection) OnRead(r func(connection *Connection, cp *ConnectionPool, frame Frame), cp *ConnectionPool) {
	// Payload
	buf := make([]byte, 1024)
	_, err := (*c.Conn).Read(buf)

	if err != nil {
		c.Close()
	}

	// Frame
	frame := Frame{}

	// Fin
	v := getBits(buf[0], 0, 1)
	if v == 1 {
		frame.IsFin = true
	} else {
		frame.IsFin = false
	}

	// Opcode
	frame.Opcode = getBits(buf[0], 4, 4)

	// Close connection
	if frame.Opcode == 8 {
		c.Close()
		return
	}

	// Ping
	if frame.Opcode == 0x9 {
		pong := buf[:]
		pong[0] = pong[0] & 0xF0
		pong[0] = pong[0] | 0x9
		(*c.Conn).Write(pong)
		return
	}

	// Is the payload masked ?
	isMasked := getBits(buf[1], 0, 1)
	frame.IsMasked = false
	if isMasked == 1 {
		frame.IsMasked = true
	}
	// Payload length
	frame.PayloadLength = uint64(getBits(buf[1], 1, 7))
	offset := 0 // In case more than one byte is used for the payload length

	if frame.PayloadLength == 126 {
		// Concatenate the next 2 bytes
		frame.PayloadLength = concatBytes(buf[2:4]...)
		offset += 2
	} else if frame.PayloadLength == 127 {
		// Concatenate the next 8 bytes
		frame.PayloadLength = concatBytes(buf[2:10]...)
		offset += 8
	}

	// Masking key
	if frame.IsMasked {
		frame.MaskingKey = [4]byte(buf[2+offset : 6+offset])
		offset += 4
	}

	// Add data
	if frame.IsMasked {
		// Unmask
		payload := buf[2+offset : uint64(2+offset)+frame.PayloadLength]
		maskUnmask(&payload, frame.MaskingKey[:])

		frame.Payload = payload
	} else {
		frame.Payload = buf[2+offset:]
	}

	r(c, cp, frame)
}

func (c *Connection) Close() {
	(*c.Conn).Write([]byte{0b10001000, 0b00000000})
	(*c.Conn).Close()
	c.isOpen = false
	return
}

// Constructs a response frame from a Frame struct
func (c *Connection) Write(f Frame) {
	bytes := []byte{0b0, 0b0}

	// IsFin
	if f.IsFin {
		bytes[0] = 0b10000000
	}

	// Opcode
	bytes[0] = bytes[0] | f.Opcode

	// Is masked
	if f.IsMasked {
		bytes[1] = bytes[1] | 0b10000000
	}

	// Payload length
	if len(f.Payload) < 126 {
		bytes[1] = bytes[1] | byte(len(f.Payload))
	} else if len(f.Payload) <= 65_535 {
		bytes[1] = bytes[1] | 126

		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(len(f.Payload)))
		bytes = append(bytes, b...)
	} else {
		bytes[1] = bytes[1] | 127

		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(len(f.Payload)))
		bytes = append(bytes, b...)
	}

	// Masking key
	if f.IsMasked {
		bytes = append(bytes, f.MaskingKey[:]...)
	}

	// Payload
	if f.IsMasked {
		// Masks the payload
		payload := f.Payload
		maskUnmask(&payload, f.MaskingKey[:])
		bytes = append(bytes, payload...)
	} else {
		bytes = append(bytes, f.Payload...)
	}

	(*c.Conn).Write(bytes)
}
