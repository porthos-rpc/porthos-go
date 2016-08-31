package server

import (
	"encoding/json"
	"errors"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

// Response represents a rpc response.
type Response struct {
	content     []byte
	contentType string
}

// ResponseWriter is responsible for sending back the response to the replyTo queue.
type ResponseWriter struct {
	channel  *amqp.Channel
	autoAck  bool
	delivery amqp.Delivery
}

// JSON sets the content of the method's response.
func (r *Response) JSON(c interface{}) {
	if c == nil {
		panic("Response content is empty")
	}

	jsonContent, err := json.Marshal(&c)

	if err != nil {
		panic(err)
	}

	r.content = jsonContent
	r.contentType = "application/json"
}

func (rw *ResponseWriter) Write(res *Response) error {
	log.Info("Sending response to queue '%s'. Slot: '%d'", rw.delivery.ReplyTo, []byte(rw.delivery.CorrelationId))

	if rw.channel == nil {
		return errors.New("No AMQP channel to publish the response to.")
	}

	err := rw.channel.Publish(
		"",
		rw.delivery.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   res.contentType,
			CorrelationId: rw.delivery.CorrelationId,
			Body:          res.content,
		})

	if err != nil {
		return err
	}

	if !rw.autoAck {
		rw.delivery.Ack(false)
		log.Debug("Ack from slot '%d'", []byte(rw.delivery.CorrelationId))
	}

	return nil
}
