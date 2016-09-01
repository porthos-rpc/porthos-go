package message

import (
	"testing"
	"unsafe"
)

func TestUintptrToBytes(t *testing.T) {
	type X struct{}
	x := X{}

	pointerInBytes := UintptrToBytes((uintptr)(unsafe.Pointer(&x)))

	if string(pointerInBytes) == "" {
		t.Errorf("Pointer was not properly converted to bytes")
	}
}

func TestBytesToUintptr(t *testing.T) {
	type X struct{ a int }
	x := X{}

	pointerInBytes := UintptrToBytes((uintptr)(unsafe.Pointer(&x)))
	pointer := BytesToUintptr(pointerInBytes)

	if pointer != (uintptr)(unsafe.Pointer(&x)) {
		t.Errorf("Invalid pointer")
	}
}

func TestBytesToUintptrInvalid(t *testing.T) {
	type X struct{ a int }

	x := X{1}
	y := X{1}

	pointerInBytesX := UintptrToBytes((uintptr)(unsafe.Pointer(&x)))
	pointerX := BytesToUintptr(pointerInBytesX)

	if pointerX == (uintptr)(unsafe.Pointer(&y)) {
		t.Errorf("Invalid pointer")
	}
}
