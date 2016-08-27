package server

import "time"

// IncomingRPC is an alias of Request.
type IncomingRPC struct {
	Request *Request
}

// OutgoingRPC contains the Request and the calculated response time.
type OutgoingRPC struct {
	Request      *Request
	Response     *Response
	ResponseTime time.Duration
}

// Extension of a porthos server.
type Extension struct {
	incoming chan IncomingRPC
	outgoing chan OutgoingRPC
}

// NewExtension creates a incoming and outgoing extension.
// You must drain both channels.
func NewExtension() *Extension {
	return &Extension{
		incoming: make(chan IncomingRPC),
		outgoing: make(chan OutgoingRPC),
	}
}

// NewIncomingExtension creates a incoming-only extension.
func NewIncomingOnlyExtension() *Extension {
	return &Extension{
		incoming: make(chan IncomingRPC),
	}
}

// NewOutgoingExtension creates a outgoing-only extension.
func NewOutgoingOnlyExtension() *Extension {
	return &Extension{
		outgoing: make(chan OutgoingRPC),
	}
}

// Incoming returns a read-only chan of incoming RPC.
func (e *Extension) Incoming() <-chan IncomingRPC {
	return e.incoming
}

// Outgoing return a read-onyl chan of outgoing RPC.
func (e *Extension) Outgoing() <-chan OutgoingRPC {
	return e.outgoing
}
