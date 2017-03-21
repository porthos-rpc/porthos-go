package porthos

import (
	"sync"
	"unsafe"
)

// Slot of a RPC call.
type Slot interface {
	// ResponseChannel returns the response channel.
	ResponseChannel() <-chan ClientResponse
	// Dispose response resources.
	Dispose()
}

type slot struct {
	responseChannel chan ClientResponse
	closed          bool
	mutex           *sync.Mutex
}

func (slot *slot) getCorrelationID() string {
	return string(UintptrToBytes((uintptr)(unsafe.Pointer(slot))))
}

func (slot *slot) ResponseChannel() <-chan ClientResponse {
	return slot.responseChannel
}

func (slot *slot) Dispose() {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if !slot.closed {
		slot.closed = true
		close(slot.responseChannel)
	}
}

func (slot *slot) sendResponse(c ClientResponse) {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if !slot.closed {
		slot.responseChannel <- c
	}
}

func NewSlot() *slot {
	return &slot{make(chan ClientResponse), false, new(sync.Mutex)}
}

func getSlot(address uintptr) *slot {
	return (*slot)(unsafe.Pointer(uintptr(address)))
}
