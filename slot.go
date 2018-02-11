package porthos

import (
	"sync"
)

// Slot of a RPC call.
type Slot interface {
	// ResponseChannel returns the response channel.
	ResponseChannel() <-chan ClientResponse
	// Dispose response resources.
	Dispose()
	// Correlation ID
	GetCorrelationID() (string, error)
}

type slot struct {
	responseChannel chan ClientResponse
	mutex           sync.Mutex
	id              string
}

func (slot *slot) GetCorrelationID() (string, error) {
	var err error

	if slot.id == "" {
		slot.id, err = NewUUIDv4()
	}

	return slot.id, err
}

func (slot *slot) ResponseChannel() <-chan ClientResponse {
	return slot.responseChannel
}

func (slot *slot) Dispose() {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if slot.responseChannel != nil {
		close(slot.responseChannel)
		slot.responseChannel = nil
	}
}

func (slot *slot) sendResponse(c ClientResponse) {
	slot.mutex.Lock()
	defer slot.mutex.Unlock()

	if slot.responseChannel != nil {
		slot.responseChannel <- c
	}
}

func NewSlot() *slot {
	return &slot{
		responseChannel: make(chan ClientResponse),
	}
}
