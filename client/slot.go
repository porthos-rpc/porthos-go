package client

import (
	"sync"
	"unsafe"

	"github.com/porthos-rpc/porthos-go/message"
)

// Slot of a RPC call.
type Slot struct {
	responseChannel chan Response
	closed          bool
	mutex           *sync.Mutex
}

func (r *Slot) getCorrelationID() string {
	return string(message.UintptrToBytes((uintptr)(unsafe.Pointer(r))))
}

// ResponseChannel returns the response channel.
func (r *Slot) ResponseChannel() <-chan Response {
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
