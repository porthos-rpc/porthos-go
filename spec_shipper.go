package porthos

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

const specsQueueName = "porthos.specs"

type specEntry struct {
	Service string          `json:"service"`
	Specs   map[string]Spec `json:"specs"`
}

// SpecShipperExtension logs incoming requests and outgoing responses.
type SpecShipperExtension struct {
	b *Broker
}

// ServerListening takes all registered method specs and ships to the broker.
func (s *SpecShipperExtension) ServerListening(srv Server) error {
	ch, err := s.b.openChannel()

	if err != nil {
		return fmt.Errorf("Error opening channel for the spec shipper. Error: %s", err)
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
		return fmt.Errorf("Error declaring the specs queue. Error: %s", err)
	}

	payload, err := json.Marshal(specEntry{
		Service: srv.GetServiceName(),
		Specs:   srv.GetSpecs(),
	})

	if err != nil {
		return fmt.Errorf("Error creating specs payload. Error: %s", err)
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
		return fmt.Errorf("Error publishing specs to the broker. Error: %s", err)
	}

	return nil
}

// IncomingRequest this is not implemented in this extension.
func (s *SpecShipperExtension) IncomingRequest(req Request) {}

// OutgoingResponse this is not implemented in this extension.
func (s *SpecShipperExtension) OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int32) {
}

// NewSpecShipperExtension creates a new extension that ship method specs to the broker.
func NewSpecShipperExtension(b *Broker) Extension {
	return &SpecShipperExtension{b}
}
