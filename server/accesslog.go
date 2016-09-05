package server

import (
	"github.com/porthos-rpc/porthos-go/log"
)

// NewAccessLogExtension creates a new extension that logs everything to stdout.
func NewAccessLogExtension() *Extension {
	ext := NewOutgoingOnlyExtension()

	go func() {
		for {
			out := <-ext.Outgoing()

			log.Info("Method [%s] Body [%s] Status Code [%d] Response Time [%fms]",
				out.Request.MethodName, out.Request.messageBody, out.Response.statusCode, out.ResponseTime.Seconds()*1000)
		}
	}()

	return ext
}
