package porthos

import (
	"github.com/streadway/amqp"
)

// ResponseWriter is responsible for sending back the response to the replyTo queue.
type ResponseWriter interface {
	Write(res Response) error
}

type responseWriter struct {
	channel  *amqp.Channel
	autoAck  bool
	delivery amqp.Delivery
}

func (rw *responseWriter) Write(res Response) error {
	if rw.channel == nil {
		return ErrNilPublishChannel
	}

	// status code is a header as well.
	res.GetHeaders().Set("statusCode", res.GetStatusCode())

	err := rw.channel.Publish(
		"",
		rw.delivery.ReplyTo,
		false,
		false,
		amqp.Publishing{
			Headers:       res.GetHeaders().asMap(),
			ContentType:   res.GetContentType(),
			CorrelationId: rw.delivery.CorrelationId,
			Body:          res.GetBody(),
		})

	if err != nil {
		return err
	}

	if !rw.autoAck {
		rw.delivery.Ack(false)
	}

	return nil
}
