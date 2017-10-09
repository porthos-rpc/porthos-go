package porthos

import (
	"errors"
)

var (
	ErrTimedOut          = errors.New("timed out")
	ErrNilPublishChannel = errors.New("No AMQP channel to publish the response to.")
	ErrNotAcked          = errors.New("Request was no acked.")
)
