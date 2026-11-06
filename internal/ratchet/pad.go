package ratchet

import (
	"encoding/binary"
	"fmt"
)

const MaxMessageSize = 16*1024 - 4

var paddingBuckets = [...]int{256, 1024, 4096, 16384}

func Pad(message []byte) ([]byte, error) {
	for _, size := range paddingBuckets {
		if len(message) <= size-4 {
			padded := make([]byte, size)
			binary.BigEndian.PutUint32(padded, uint32(len(message)))
			copy(padded[4:], message)
			return padded, nil
		}
	}
	return nil, fmt.Errorf("message exceeds %d bytes", MaxMessageSize)
}

func Unpad(padded []byte) ([]byte, error) {
	validSize := false
	for _, size := range paddingBuckets {
		validSize = validSize || len(padded) == size
	}
	if !validSize {
		return nil, fmt.Errorf("invalid padding bucket")
	}
	length := binary.BigEndian.Uint32(padded[:4])
	if length > uint32(len(padded)-4) {
		return nil, fmt.Errorf("invalid padded message length")
	}
	for _, b := range padded[4+int(length):] {
		if b != 0 {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return padded[4 : 4+int(length)], nil
}
