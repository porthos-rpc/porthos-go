package porthos

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// UintptrToBytes converts a pointer to a bytes array.
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

// BytesToUintptr converts a bytes array to a pointer.
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
