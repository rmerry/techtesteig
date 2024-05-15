package btcore

import (
	"crypto/sha256"
	"encoding/binary"
)

// If payload is empty, as in verack and “getaddr” messages, the checksum is
// always 0x5df6e0e2 (SHA256(SHA256(<empty string>))).
const emptyPayloadChecksum = 0x5df6e0e2

func checksum(input []byte) uint32 {
	if input == nil || len(input) == 0 {
		return emptyPayloadChecksum
	}

	h1 := sha256.New()
	h2 := sha256.New()

	h1.Write(input)
	h2.Write(h1.Sum(nil))

	return binary.LittleEndian.Uint32(h2.Sum(nil)[:4])
}
