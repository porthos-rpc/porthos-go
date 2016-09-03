package server

import (
	"os"
	"testing"
	"time"
)

func TestMetricsShipperExtension(t *testing.T) {
	broker, _ := NewBroker(os.Getenv("AMQP_URL"))

	ext := NewMetricsShipperExtension(broker, MetricsShipperConfig{BufferSize: 2})

	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test1"}, &Response{}, 4 * time.Millisecond, 200}
	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test2"}, &Response{}, 5 * time.Millisecond, 201}
	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test2"}, &Response{}, 6 * time.Millisecond, 201}
	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test3"}, &Response{}, 7 * time.Millisecond, 202}
	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test4"}, &Response{}, 8 * time.Millisecond, 200}
	ext.outgoing <- OutgoingRPC{&Request{MethodName: "test4"}, &Response{}, 9 * time.Millisecond, 200}

	ch, _ := broker.conn.Channel()

	dc, _ := ch.Consume(
		"porthos.metrics", // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)

	shippedMetricsCount := 0

	go func() {
		for _ = range dc {
			shippedMetricsCount++
		}
	}()

	<-time.After(2 * time.Second)

	if shippedMetricsCount != 3 {
		t.Errorf("Excepted 3 shipped metrics, got %d", shippedMetricsCount)
	}
}
