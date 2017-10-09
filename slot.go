package porthos

import (
	"github.com/satori/go.uuid"
	"sync"
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
	mutex           *sync.Mutex
}

func (slot *slot) getCorrelationID() string {
	uid := uuid.NewV4()
	return uid.String()
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
	return &slot{make(chan ClientResponse), new(sync.Mutex)}
}
