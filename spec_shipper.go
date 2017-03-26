package porthos

import (
	"encoding/json"
	"time"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

const specsQueueName = "porthos.specs"

// SpecShipperExtension logs incoming requests and outgoing responses.
type SpecShipperExtension struct {
	b *Broker
}

// ServerListening takes all registered method specs and ships to the broker.
func (s *SpecShipperExtension) ServerListening(srv Server) {
	ch, err := s.b.openChannel()

	if err != nil {
		log.Error("Error opening channel for the spec shipper.")
		return
	}

	defer ch.Close()

	_, err = ch.QueueDeclare(
		specsQueueName, // name
		true,           // durable
		false,          // delete when usused
		false,          // exclusive
		false,          // noWait
		nil,            // arguments
	)

	if err != nil {
		log.Error("Error declaring the specs queue", err)
	}

	payload, err := json.Marshal(srv.(*server).specs)

	if err != nil {
		log.Error("Error creating specs payload", err)
	}

	err = ch.Publish(
		"",
		specsQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})

	if err != nil {
		log.Error("Error publishing specs to the broker", err)
	}
}

// IncomingRequest this is not implemented in this extension.
func (s *SpecShipperExtension) IncomingRequest(req Request) {}

// OutgoingResponse this is not implemented in this extension.
func (s *SpecShipperExtension) OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int16) {
}

// NewSpecShipperExtension creates a new extension that ship method specs to the broker.
func NewSpecShipperExtension(b *Broker) Extension {
	return &SpecShipperExtension{b}
}
