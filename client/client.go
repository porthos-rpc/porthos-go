package client

import (
	"sync"
	"time"
	"unsafe"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

// Client is an entry point for making remote calls.
type Client struct {
	serviceName     string
	defaultTTL      time.Duration
	channel         *amqp.Channel
	deliveryChannel <-chan amqp.Delivery
	responseQueue   *amqp.Queue
}

// NewClient creates a new instance of Client, responsible for making remote calls.
func NewClient(b *broker.Broker, serviceName string, defaultTTL time.Duration) (*Client, error) {
	ch, err := b.OpenChannel()

	if err != nil {
		return nil, err
	}

	// create the response queue (let the amqp server to pick a name for us)
	q, err := ch.QueueDeclare("", false, false, true, false, nil)

	if err != nil {
		ch.Close()
		return nil, err
	}

	dc, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		ch.Close()
		return nil, err
	}

	c := &Client{
		serviceName,
		defaultTTL,
		ch,
		dc,
		&q,
	}

	c.start()

	return c, nil
}

func (c *Client) start() {
	go func() {
		for d := range c.deliveryChannel {
			c.processResponse(d)
		}
	}()
}

func (c *Client) processResponse(d amqp.Delivery) {
	d.Ack(false)

	log.Debug("Ack. Received response in '%s' for slot: '%d'", d.RoutingKey, []byte(d.CorrelationId))

	address := c.unmarshallCorrelationID(d.CorrelationId)

	res := c.getSlot(address)

	func() {
		res.mutex.Lock()
		defer res.mutex.Unlock()

		if !res.closed {
			res.responseChannel <- Response{
				Content:     d.Body,
				ContentType: d.ContentType,
				StatusCode:  d.Headers["statusCode"].(int16),
				Headers:     d.Headers,
			}
		}
	}()
}

// Call prepares a remote call.
func (c *Client) Call(method string) *call {
	return newCall(c, method)
}

// Close the client and AMQP chanel.
func (c *Client) Close() {
	c.channel.Close()
}

func (c *Client) getSlot(address uintptr) *Slot {
	return (*Slot)(unsafe.Pointer(uintptr(address)))
}

func (c *Client) makeNewSlot() *Slot {
	return &Slot{make(chan Response), false, new(sync.Mutex)}
}

func (c *Client) unmarshallCorrelationID(correlationID string) uintptr {
	return BytesToUintptr([]byte(correlationID))
}
