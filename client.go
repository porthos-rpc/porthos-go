package porthos

import (
	"time"

	"github.com/streadway/amqp"
)

// Client is an entry point for making remote calls.
type Client struct {
	serviceName     string
	defaultTTL      time.Duration
	channel         *amqp.Channel
	deliveryChannel <-chan amqp.Delivery
	responseQueue   *amqp.Queue
	broker          *Broker
}

// NewClient creates a new instance of Client, responsible for making remote calls.
func NewClient(b *Broker, serviceName string, defaultTTL time.Duration) (*Client, error) {
	ch, err := b.openChannel()

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
		b,
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

	address := c.unmarshallCorrelationID(d.CorrelationId)

	statusCode := d.Headers["statusCode"].(int32)

	res := getSlot(address)
	res.sendResponse(ClientResponse{
		Content:     d.Body,
		ContentType: d.ContentType,
		StatusCode:  statusCode,
		Headers:     *NewHeadersFromMap(d.Headers),
	})
}

// Call prepares a remote call.
func (c *Client) Call(method string) *call {
	return newCall(c, method)
}

// Close the client and AMQP chanel.
func (c *Client) Close() {
	c.channel.Close()
}

func (c *Client) unmarshallCorrelationID(correlationID string) uintptr {
	return BytesToUintptr([]byte(correlationID))
}
