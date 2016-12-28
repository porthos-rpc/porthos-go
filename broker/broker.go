package broker

import (
	"errors"

	"github.com/streadway/amqp"
)

// Broker holds an implementation-specific connection.
type Broker struct {
	Conn *amqp.Connection
}

// Error holds an implementation-specific error.
type Error struct {
	brokerError interface{}
}

// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*Broker, error) {
	conn, err := amqp.Dial(amqpURL)

	if err != nil {
		return nil, err
	}

	return &Broker{conn}, nil
}

// Close the broker connection.
func (b *Broker) Close() {
	b.Conn.Close()
}

// NotifyConnectionClose writes in the returned channel when the connection with the broker closes.
func (b *Broker) NotifyConnectionClose() <-chan error {
	ch := make(chan error)

	go func() {
		ch <- errors.New((<-b.Conn.NotifyClose(make(chan *amqp.Error))).Error())
	}()

	return ch
}

// WaitUntilConnectionCloses hold the current goroutine until the connection with the broker closes.
func (b *Broker) WaitUntilConnectionCloses() {
	<-b.NotifyConnectionClose()
}
