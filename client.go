package porthos

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

// Client is an entry point for making remote calls.
type Client struct {
	serviceName     string
	defaultTTL      time.Duration
	channel         *amqp.Channel
	deliveryChannel <-chan amqp.Delivery
	responseQueue   *amqp.Queue
	broker          *Broker
	slots           map[string]*slot
	slotLock        *sync.Mutex
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
		serviceName:     serviceName,
		defaultTTL:      defaultTTL,
		channel:         ch,
		deliveryChannel: dc,
		responseQueue:   &q,
		broker:          b,
		slots:           make(map[string]*slot, 3000),
		slotLock:        new(sync.Mutex),
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

	statusCode := d.Headers["statusCode"].(int32)

	res, ok := c.popSlot(d.CorrelationId)
	if ok {
		res.sendResponse(ClientResponse{
			Content:     d.Body,
			ContentType: d.ContentType,
			StatusCode:  statusCode,
			Headers:     *NewHeadersFromMap(d.Headers),
		})
	} else {
		log.Printf("Slot %s not exists.", d.CorrelationId)
	}
}

// Call prepares a remote call.
func (c *Client) Call(method string) *call {
	return newCall(c, method)
}

// Close the client and AMQP chanel.
func (c *Client) Close() {
	c.channel.Close()
}

func (c *Client) pushSlot(correlationID string, slot *slot) {
	c.slotLock.Lock()
	defer c.slotLock.Unlock()

	c.slots[correlationID] = slot
}

func (c *Client) popSlot(correlationID string) (*slot, bool) {
	c.slotLock.Lock()
	defer c.slotLock.Unlock()

	slot, ok := c.slots[correlationID]

	if ok {
		delete(c.slots, correlationID)
	}

	return slot, ok
}
