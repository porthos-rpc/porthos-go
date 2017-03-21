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
	// Retuns the correlation id
	GetCorrelationID() string
	// Send ClientResponse to slot
	SendResponse(c ClientResponse)
}

type SlotImpl struct {
	responseChannel chan ClientResponse
	closed          bool
	mutex           *sync.Mutex
}

func (slot *SlotImpl) GetCorrelationID() string {
	return string(UintptrToBytes((uintptr)(unsafe.Pointer(slot))))
}

func (slot *SlotImpl) ResponseChannel() <-chan ClientResponse {
	return slot.responseChannel
}

func (slot *SlotImpl) Dispose() {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if !slot.closed {
		slot.closed = true
		close(slot.responseChannel)
	}
}

func (slot *SlotImpl) SendResponse(c ClientResponse) {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if !slot.closed {
		slot.responseChannel <- c
	}
}

func NewSlot() Slot {
	return &SlotImpl{make(chan ClientResponse), false, new(sync.Mutex)}
}

func getSlot(address uintptr) Slot {
	return (*SlotImpl)(unsafe.Pointer(uintptr(address)))
}
