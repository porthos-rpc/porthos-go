package porthos

import (
	"sync"
	"unsafe"
)

// Slot of a RPC call.
type Slot struct {
	responseChannel chan ClientResponse
	closed          bool
	mutex           *sync.Mutex
}

func (r *Slot) getCorrelationID() string {
	return string(UintptrToBytes((uintptr)(unsafe.Pointer(r))))
}

// ResponseChannel returns the response channel.
func (r *Slot) ResponseChannel() <-chan ClientResponse {
	return r.responseChannel
}

// Dispose response resources.
func (r *Slot) Dispose() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.closed {
		r.closed = true
		close(r.responseChannel)
	}
}
