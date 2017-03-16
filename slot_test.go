package porthos

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestGetCorrelationId(t *testing.T) {
	slot := Slot{}

	pointerStr := slot.getCorrelationID()

	var intptr uintptr
	size := unsafe.Sizeof(intptr)
	switch size {
	case 4:
		intptr = uintptr(binary.LittleEndian.Uint32([]byte(pointerStr)))
	case 8:
		intptr = uintptr(binary.LittleEndian.Uint64([]byte(pointerStr)))
	}

	if intptr != (uintptr)(unsafe.Pointer(&slot)) {
		t.Errorf("Invalid correlation Id")
	}
}
