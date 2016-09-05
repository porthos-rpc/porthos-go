package server

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

type ResponseExample struct {
	Sum float64 `json:"sum"`
}

func TestResponseWriterJSON(t *testing.T) {
	broker, _ := NewBroker(os.Getenv("AMQP_URL"))
	ch, _ := broker.conn.Channel()
	q, _ := ch.QueueDeclare("", false, false, true, false, nil)
	dc, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	response := NewResponse()
	response.JSON(200, ResponseExample{Sum: 10})

	rw := ResponseWriter{
		channel: ch,
		autoAck: true,
		delivery: amqp.Delivery{
			ReplyTo:       q.Name,
			CorrelationId: "correlationId",
		},
	}

	rw.Write(response)

	select {
	case response := <-dc:
		if response.Headers["statusCode"].(int16) != 200 {
			t.Errorf("Expected status code was 200, got: %d", response.Headers["statusCode"])
		}

		if response.ContentType != "application/json" {
			t.Errorf("Expected content type application/json, got: %s", response.ContentType)
		}

		var responseTest ResponseExample
		json.Unmarshal(response.Body, &responseTest)

		if responseTest.Sum != 10 {
			t.Errorf("Response failed, expected: 10, got: %s", responseTest.Sum)
		}

		return
	case <-time.After(3 * time.Second):
		t.Fatal("No response receive. Timedout.")
	}
}

func TestResponseWriterRaw(t *testing.T) {
	broker, _ := NewBroker(os.Getenv("AMQP_URL"))
	ch, _ := broker.conn.Channel()
	q, _ := ch.QueueDeclare("", false, false, true, false, nil)
	dc, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	response := NewResponse()
	response.Raw(201, "text/plain", []byte("Some Response Text"))

	rw := ResponseWriter{
		channel: ch,
		autoAck: true,
		delivery: amqp.Delivery{
			ReplyTo:       q.Name,
			CorrelationId: "correlationId",
		},
	}

	rw.Write(response)

	select {
	case response := <-dc:
		if response.Headers["statusCode"].(int16) != 201 {
			t.Errorf("Expected status code was 201, got: %d", response.Headers["statusCode"])
		}

		if response.ContentType != "text/plain" {
			t.Errorf("Expected content type text/plain, got: %s", response.ContentType)
		}

		if string(response.Body) != "Some Response Text" {
			t.Errorf("Response failed, expected: Some Response Text, got: %s", string(response.Body))
		}

		return
	case <-time.After(3 * time.Second):
		t.Fatal("No response receive. Timedout.")
	}
}

func TestResponseWriterEmpty(t *testing.T) {
	broker, _ := NewBroker(os.Getenv("AMQP_URL"))
	ch, _ := broker.conn.Channel()
	q, _ := ch.QueueDeclare("", false, false, true, false, nil)
	dc, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	response := NewResponse()
	response.Empty(202)

	rw := ResponseWriter{
		channel: ch,
		autoAck: true,
		delivery: amqp.Delivery{
			ReplyTo:       q.Name,
			CorrelationId: "correlationId",
		},
	}

	rw.Write(response)

	select {
	case response := <-dc:
		if response.Headers["statusCode"].(int16) != 202 {
			t.Errorf("Expected status code was 202, got: %d", response.Headers["statusCode"])
		}

		if response.ContentType != "" {
			t.Errorf("Expected content type nil, got: %s", response.ContentType)
		}

		if string(response.Body) != "" {
			t.Errorf("Response failed, expected: nil, got: %s", string(response.Body))
		}

		return
	case <-time.After(3 * time.Second):
		t.Fatal("No response receive. Timedout.")
	}
}
