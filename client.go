package porthos

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Client is an entry point for making remote calls.
type Client struct {
	serviceName       string
	defaultTTL        time.Duration
	broker            *Broker
	responseQueueName string

	slots    map[string]*slot
	slotLock sync.Mutex

	m      sync.Mutex
	closed bool
}

func newUniqueQueueName(prefix string) string {
	return fmt.Sprintf("%s@%d-porthos", prefix, time.Now().UnixNano())
}

// NewClient creates a new instance of Client, responsible for making remote calls.
func NewClient(b *Broker, serviceName string, defaultTTL time.Duration) (*Client, error) {
	c := &Client{
		serviceName:       serviceName,
		defaultTTL:        defaultTTL,
		broker:            b,
		slots:             make(map[string]*slot, 3000),
		responseQueueName: newUniqueQueueName(serviceName),
	}

	go c.start()

	return c, nil
}

func (c *Client) start() {
	rs := c.broker.NotifyReestablish()

	for !c.closed {
		if !c.broker.IsConnected() {
			log.Printf("[PORTHOS] Connection not established. Waiting connection to be reestablished.")

			<-rs

			continue
		}

		err := c.consume()

		if err != nil {
			log.Printf("[PORTHOS] Error consuming responses. Error: %s", err)
		} else {
			log.Print("[PORTHOS] Consuming stopped.")
		}
	}
}

func (c *Client) consume() error {
	ch, err := c.broker.openChannel()

	if err != nil {
		return err
	}

	defer ch.Close()

	// create the response queue
	_, err = ch.QueueDeclare(
		c.responseQueueName, // name
		false,               // durable
		false,               // auto-delete
		true,                // exclusive
		false,               // no-wait
		nil,
	)

	if err != nil {
		return err
	}

	dc, err := ch.Consume(
		c.responseQueueName, // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)

	if err != nil {
		return err
	}

	for d := range dc {
		c.processResponse(d)
	}

	return nil
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
		log.Printf("[PORTHOS] Slot %s not exists.", d.CorrelationId)
	}
}

// Call prepares a remote call.
func (c *Client) Call(method string) *call {
	return newCall(c, method)
}

// Close the client and AMQP chanel.
// Client only will die if broker was closed.
func (c *Client) Close() {
	c.m.Lock()
	defer c.m.Unlock()
	c.closed = true
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
