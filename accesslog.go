package porthos

import (
	"time"

	"github.com/porthos-rpc/porthos-go/log"
)

// AccessLogExtension logs incoming requests and outgoing responses.
type AccessLogExtension struct {
}

// ServerListening this is not implemented in this extension.
func (a *AccessLogExtension) ServerListening(server Server) {}

// IncomingRequest logs rpc request method and arguments.
func (a *AccessLogExtension) IncomingRequest(req Request) {
	log.Info("Method [%s] Arguments [%s]",
		req.GetMethodName(),
		string(req.GetBody()))
}

// OutgoingResponse logs rpc response details.
func (a *AccessLogExtension) OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int16) {
	log.Info("Method [%s] Arguments [%s] Status Code [%d] Response Time [%fms]",
		req.GetMethodName(),
		string(req.GetBody()),
		statusCode,
		resTime.Seconds()*1000)
}

// NewAccessLogExtension creates a new extension that logs everything to stdout.
func NewAccessLogExtension() Extension {
	return &AccessLogExtension{}
}
