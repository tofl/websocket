package websocket

// Returns c bits at position p from left to right
func getBits(b byte, p, c int8) byte {
	return (b >> (8 - (p + c))) & (0b11111111 >> (8 - c))
}

func concatBytes(bytes ...byte) uint64 {
	var finalNumber uint64
	for i := 0; i < len(bytes); i++ {
		finalNumber = (finalNumber << 8) | uint64(bytes[i])
	}

	return finalNumber
}

func maskUnmask(data *[]byte, key []byte) {
	for k, v := range *data {
		(*data)[k] = v ^ key[k%len(key)]
	}
}
