package porthos

import (
	"log"
	"time"
)

// AccessLogExtension logs incoming requests and outgoing responses.
type AccessLogExtension struct {
}

// ServerListening this is not implemented in this extension.
func (a *AccessLogExtension) ServerListening(server Server) error {
	return nil
}

// IncomingRequest logs rpc request method and arguments.
func (a *AccessLogExtension) IncomingRequest(req Request) {
	log.Printf("[PORTHOS] Method [%s] Arguments [%s]",
		req.GetMethodName(),
		string(req.GetBody()))
}

// OutgoingResponse logs rpc response details.
func (a *AccessLogExtension) OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int32) {
	log.Printf("[PORTHOS] Method [%s] Arguments [%s] Status Code [%d] Response Time [%fms]",
		req.GetMethodName(),
		string(req.GetBody()),
		statusCode,
		resTime.Seconds()*1000)
}

// NewAccessLogExtension creates a new extension that logs everything to stdout.
func NewAccessLogExtension() Extension {
	return &AccessLogExtension{}
}
