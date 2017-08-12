package porthos

import (
	"sync"
	"time"

	"github.com/streadway/amqp"
	"log"
)

// Broker holds an implementation-specific connection.
type Broker struct {
	config       Config
	connection   *amqp.Connection
	m            sync.Mutex
	url          string
	closed       bool
	reestablishs []chan bool
}

// Config to be used when creating a new connection.
type Config struct {
	reconnectInterval time.Duration
}

// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*Broker, error) {
	return NewBrokerConfig(amqpURL, Config{
		reconnectInterval: 1 * time.Second,
	})
}

// NewBrokerConfig returns an AMQP Connection.
func NewBrokerConfig(amqpURL string, config Config) (*Broker, error) {
	conn, err := amqp.Dial(amqpURL)

	if err != nil {
		return nil, err
	}

	b := &Broker{
		connection: conn,
		url:        amqpURL,
		config:     config,
	}

	go b.handleConnectionClose()

	return b, nil
}

// Close the broker connection.
func (b *Broker) Close() {
	b.m.Lock()
	defer b.m.Unlock()

	b.connection.Close()
}

// NotifyConnectionClose writes in the returned channel when the connection with the broker closes.
func (b *Broker) NotifyConnectionClose() <-chan error {
	ch := make(chan error)

	go func() {
		ch <- <-b.connection.NotifyClose(make(chan *amqp.Error))
	}()

	return ch
}

// NotifyReestablish returns a channel to notify when the connection is restablished.
func (b *Broker) NotifyReestablish() <-chan bool {
	receiver := make(chan bool)
	b.reestablishs = append(b.reestablishs, receiver)

	return receiver
}

func (b *Broker) openChannel() (*amqp.Channel, error) {
	b.m.Lock()
	defer b.m.Unlock()

	return b.connection.Channel()
}

func (b *Broker) reestablish() error {
	conn, err := amqp.Dial(b.url)

	b.m.Lock()
	defer b.m.Unlock()

	b.connection = conn

	return err

}

func (b *Broker) handleConnectionClose() {
	for !b.closed {
		<-b.NotifyConnectionClose()

		for i := 0; !b.closed; i++ {
			err := b.reestablish()

			if err == nil {
				log.Printf("[PORTHOS] Connection reestablished")

				for _, c := range b.reestablishs {
					c <- true
				}

				break
			} else {
				log.Printf("[PORTHOS] Error reestablishing connection, attempt %d. Retrying... [%s]", i, err)

				time.Sleep(b.config.reconnectInterval)
			}
		}
	}
}
