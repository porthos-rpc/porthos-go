package porthos

import (
	"github.com/porthos-rpc/porthos-go/log"
)

// NewAccessLogExtension creates a new extension that logs everything to stdout.
func NewAccessLogExtension() *Extension {
	ext := NewOutgoingOnlyExtension()

	go func() {
		for {
			out := <-ext.Outgoing()

			log.Info("Method [%s] Arguments [%s] Status Code [%d] Response Time [%fms]",
				out.Request.GetMethodName(), string(out.Request.GetBody()), out.Response.GetStatusCode(), out.ResponseTime.Seconds()*1000)
		}
	}()

	return ext
}
