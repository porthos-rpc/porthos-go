package porthos

import (
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestNewClient(t *testing.T) {
	b, err := NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		t.Errorf("Got an error creating the broker: %s", err)
		panic(err)
	}

	defer b.Close()

	c, err := NewClient(b, "TestService", 120)

	if err != nil {
		t.Errorf("Got an error creating the client: %s", err)
		panic(err)
	}

	if c.serviceName != "TestService" {
		t.Errorf("Got an unexpected serviceName: %s", c.serviceName)
	}

	if c.defaultTTL != 120 {
		t.Errorf("Got an unexpected defaultTTL: %d", c.defaultTTL)
	}

	if c.channel == nil {
		t.Errorf("Got an nil pointer of channel")
	}

	if c.deliveryChannel == nil {
		t.Errorf("Got an nil pointer of delivery channel")
	}

	if c.responseQueue == nil {
		t.Errorf("Got an nil pointer of responseQueue")
	}

	if c.responseQueue.Name == "" {
		t.Errorf("Got an empty responseQueue name")
	}

	if c.broker == nil {
		t.Errorf("Got an nil pointer of broker")
	}

	defer c.Close()
}

func TestProcessResponse(t *testing.T) {
	b, err := NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		t.Errorf("Got an error creating the broker: %s", err)
		panic(err)
	}

	defer b.Close()

	c, err := NewClient(b, "TestService", 1500)

	if err != nil {
		t.Errorf("Got an error creating the client: %s", err)
		panic(err)
	}

	defer c.Close()

	ret, err := c.Call("simpleMethod").Async()

	if err != nil {
		t.Errorf("Got an error making the call: %s", err)
		panic(err)
	}

	defer ret.Dispose()

	err = c.channel.Publish(
		"",
		c.responseQueue.Name,
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"statusCode": int16(200),
			},
			CorrelationId: ret.(*slot).getCorrelationID(),
			Body:          []byte(""),
			ContentType:   "application/octet-stream",
		})

	if err != nil {
		t.Errorf("Got an error publishing a fake response: %s", err)
		panic(err)
	}

	select {
	case res := <-ret.ResponseChannel():
		if res.StatusCode != 200 {
			t.Errorf("Got an unexpected status code: %d", res.StatusCode)
		}

		if res.Headers["statusCode"].(int16) != 200 {
			t.Errorf("Got an unexpected status code in headers: %d", res.Headers["statusCode"].(int16))
		}

		if res.ContentType != "application/octet-stream" {
			t.Errorf("Got an unexpected contentType: %d", res.ContentType)
		}

		return
	case <-time.After(3 * time.Second):
		t.Errorf("Got an error creating the client: %s", err)
	}

}

func TestProcessResponseMultipleClients(t *testing.T) {
	b, err := NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		t.Errorf("Got an error creating the broker: %s", err)
		panic(err)
	}

	defer b.Close()

	c1, err := NewClient(b, "TestService", 1500)

	if err != nil {
		t.Errorf("Got an error creating the client: %s", err)
		panic(err)
	}

	defer c1.Close()

	c2, err := NewClient(b, "AnotherTestService", 1500)

	if err != nil {
		t.Errorf("Got an error creating the client: %s", err)
		panic(err)
	}

	defer c2.Close()

	ret1, err := c1.Call("simpleMethod").Async()

	if err != nil {
		t.Errorf("Got an error making the call: %s", err)
		panic(err)
	}

	defer ret1.Dispose()

	ret2, err := c2.Call("simpleMethod").Async()

	defer ret2.Dispose()

	err = c1.channel.Publish(
		"",
		c1.responseQueue.Name,
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"statusCode": int16(200),
			},
			CorrelationId: ret1.(*slot).getCorrelationID(),
			Body:          []byte(""),
			ContentType:   "application/octet-stream",
		})

	if err != nil {
		t.Errorf("Got an error publishing a fake response: %s", err)
		panic(err)
	}

	err = c1.channel.Publish(
		"",
		c2.responseQueue.Name,
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"statusCode": int16(404),
			},
			CorrelationId: ret2.(*slot).getCorrelationID(),
			Body:          []byte(""),
			ContentType:   "application/octet-stream",
		})

	if err != nil {
		t.Errorf("Got an error publishing a fake response: %s", err)
		panic(err)
	}

	select {
	case res := <-ret1.ResponseChannel():
		if res.StatusCode != 200 {
			t.Errorf("Got an unexpected status code: %d", res.StatusCode)
		}

		if res.Headers["statusCode"].(int16) != 200 {
			t.Errorf("Got an unexpected status code in headers: %d", res.Headers["statusCode"].(int16))
		}

		if res.ContentType != "application/octet-stream" {
			t.Errorf("Got an unexpected contentType: %d", res.ContentType)
		}

		return
	case res := <-ret2.ResponseChannel():
		if res.StatusCode != 404 {
			t.Errorf("Got an unexpected status code: %d", res.StatusCode)
		}

		if res.Headers["statusCode"].(int16) != 404 {
			t.Errorf("Got an unexpected status code in headers: %d", res.Headers["statusCode"].(int16))
		}

		if res.ContentType != "application/octet-stream" {
			t.Errorf("Got an unexpected contentType: %d", res.ContentType)
		}

		return
	case <-time.After(3 * time.Second):
		t.Errorf("Got an error creating the client: %s", err)
	}

}
