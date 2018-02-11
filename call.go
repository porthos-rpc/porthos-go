package porthos

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type call struct {
	client      *Client
	timeout     time.Duration
	method      string
	body        []byte
	contentType string
}

// Map is an abstraction for map[string]interface{} to be used with WithMap.
type Map map[string]interface{}

// NewCall creates a new RPC call object.
func newCall(client *Client, method string) *call {
	return &call{client: client, method: method}
}

// WithTimeout defines the timeut for this specific call.
func (c *call) WithTimeout(timeout time.Duration) *call {
	c.timeout = timeout
	return c
}

// WithBody defines the given bytes array as the request body.
func (c *call) WithBody(body []byte) *call {
	c.body = body
	c.contentType = "application/octet-stream"

	return c
}

// WithBody defines the given bytes array as the request body.
func (c *call) WithBodyContentType(body []byte, contentType string) *call {
	c.body = body
	c.contentType = contentType

	return c
}

// WitArgs defines the given args as the request body.
func (c *call) WithArgs(args ...interface{}) *call {
	return c.withJSON(args)
}

// WithMap defines the given map as the request body.
func (c *call) WithMap(m map[string]interface{}) *call {
	return c.withJSON(m)
}

// WithStruct defines the given struct as the request body.
func (c *call) WithStruct(i interface{}) *call {
	return c.withJSON(i)
}

func (c *call) withJSON(i interface{}) *call {
	data, err := json.Marshal(i)

	if err != nil {
		panic(err)
	}

	c.body = data
	c.contentType = "application/json"

	return c
}

// Async calls the remote method with the given arguments.
// It returns a *Slot (which contains the response channel) and any possible error.
func (c *call) Async() (Slot, error) {
	if !c.client.broker.IsConnected() {
		return nil, ErrBrokerNotConnected
	}

	res := NewSlot()
	correlationID, err := res.GetCorrelationID()

	if err != nil {
		return nil, err
	}

	c.client.pushSlot(correlationID, res)

	ch, err := c.client.broker.openChannel()

	if err != nil {
		return nil, err
	}

	defer ch.Close()

	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("Channel could not be put into confirm mode: %s", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	err = ch.Publish(
		"",                   // exchange
		c.client.serviceName, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			Headers: amqp.Table{
				"X-Method": c.method,
			},
			Expiration:    strconv.FormatInt(c.getTimeoutMilliseconds(), 10),
			ContentType:   c.contentType,
			CorrelationId: correlationID,
			ReplyTo:       c.client.responseQueueName,
			Body:          c.body,
		})

	if err != nil {
		return nil, err
	} else {
		if confirmed := <-confirms; !confirmed.Ack {
			return nil, ErrNotAcked
		}
	}

	return res, nil
}

// Sync calls the remote method with the given arguments.
// It returns a Response and any possible error.
func (c *call) Sync() (*ClientResponse, error) {
	slot, err := c.Async()

	if err != nil {
		return nil, err
	}

	defer slot.Dispose()

	select {
	case response := <-slot.ResponseChannel():
		return &response, nil
	case <-time.After(c.getTimeout()):
		return nil, ErrTimedOut
	}
}

// Void calls a remote service procedure/service which will not provide any return value.
func (c *call) Void() error {
	if !c.client.broker.IsConnected() {
		return ErrBrokerNotConnected
	}

	ch, err := c.client.broker.openChannel()

	if err != nil {
		return err
	}

	defer ch.Close()

	err = ch.Publish(
		"",                   // exchange
		c.client.serviceName, // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			Headers: amqp.Table{
				"X-Method": c.method,
			},
			ContentType: c.contentType,
			Body:        c.body,
		})

	if err != nil {
		return err
	}

	return nil
}

func (c *call) getTimeout() time.Duration {
	if c.timeout > 0 {
		return c.timeout
	}

	return c.client.defaultTTL
}

func (c *call) getTimeoutMilliseconds() int64 {
	t := c.client.defaultTTL

	if c.timeout > 0 {
		t = c.timeout
	}

	return int64(t / time.Millisecond)
}
