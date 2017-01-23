package client

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/errors"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/message"
	"github.com/streadway/amqp"
)

// Client is an entry point for making remote calls.
type Client struct {
	serviceName     string
	defaultTTL      int64
	channel         *amqp.Channel
	deliveryChannel <-chan amqp.Delivery
	responseQueue   *amqp.Queue
}

// NewClient creates a new instance of Client, responsible for making remote calls.
func NewClient(b *broker.Broker, serviceName string, defaultTTL int64) (*Client, error) {
	ch, err := b.Conn.Channel()

	if err != nil {
		b.Conn.Close()
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

// Call calls the remote method with the given arguments.
// It returns a *Slot (which contains the response channel) and any possible error.
func (c *Client) Call(method string, args ...interface{}) (*Slot, error) {
	body, err := json.Marshal(&message.MessageBody{method, args})

	if err != nil {
		return nil, err
	}

	res := c.makeNewSlot()
	correlationID := res.getCorrelationID()

	err = c.channel.Publish(
		"",            // exchange
		c.serviceName, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			Expiration:    strconv.FormatInt(c.defaultTTL, 10),
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       c.responseQueue.Name,
			Body:          body,
		})

	log.Info("Published method '%s' in '%s'. Expecting response in queue '%s' and slot '%d'", method, c.serviceName, c.responseQueue.Name, []byte(correlationID))

	if err != nil {
		return nil, err
	}

	return res, nil
}

// CallSync calls the remote method with the given arguments.
// It returns a Response and any possible error.
func (c *Client) CallSync(method string, timeout time.Duration, args ...interface{}) (*Response, error) {
	slot, err := c.Call(method, args...)

	if err != nil {
		return nil, err
	}

	defer slot.Dispose()

	select {
	case response := <-slot.ResponseChannel():
		return &response, nil
	case <-time.After(timeout):
		return nil, errors.ErrTimedOut
	}
}

// CallVoid calls a remote service procedure/service which will not provide any return value.
func (c *Client) CallVoid(method string, args ...interface{}) error {
	body, err := json.Marshal(&message.MessageBody{method, args})

	if err != nil {
		return err
	}

	err = c.channel.Publish(
		"",            // exchange
		c.serviceName, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return err
	}

	return nil
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
	return message.BytesToUintptr([]byte(correlationID))
}
