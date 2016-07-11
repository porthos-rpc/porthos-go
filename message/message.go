package message

import (
	"encoding/binary"
	"unsafe"
	"fmt"
)

type MessageBody struct {
    Method  string          `json:"method"`
    Args    []interface{}   `json:"args"`
}

func UintptrToBytes(value uintptr) []byte {
	size := unsafe.Sizeof(value)
	b := make([]byte, size)
	switch size {
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(b, uint64(value))
	default:
		panic(fmt.Sprintf("unknown uintptr size: %v", size))
	}
	return b
}

func BytesToUintptr(bytes []byte) uintptr {
	var intptr uintptr
	size := unsafe.Sizeof(intptr)
	switch size {
	case 4:
		intptr = uintptr(binary.LittleEndian.Uint32(bytes))
	case 8:
		intptr = uintptr(binary.LittleEndian.Uint64(bytes))
	default:
		panic(fmt.Sprintf("unknown uintptr size: %v", size))
	}
	return intptr
}