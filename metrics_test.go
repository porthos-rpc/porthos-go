package porthos

import (
	"os"
	"testing"
	"time"
)

func TestMetricsShipperExtension(t *testing.T) {
	b, _ := NewBroker(os.Getenv("AMQP_URL"))

	ext, err := NewMetricsShipperExtension(b, MetricsShipperConfig{BufferSize: 2})

	if err != nil {
		t.Error(err)
	}

	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test1"}, &response{}, 4*time.Millisecond, 200)
	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test2"}, &response{}, 5*time.Millisecond, 201)
	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test2"}, &response{}, 6*time.Millisecond, 201)
	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test3"}, &response{}, 7*time.Millisecond, 202)
	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test4"}, &response{}, 8*time.Millisecond, 200)
	ext.OutgoingResponse(&request{serviceName: "SampleService", methodName: "test4"}, &response{}, 9*time.Millisecond, 200)

	ch, _ := b.openChannel()

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
		for range dc {
			shippedMetricsCount++
		}
	}()

	<-time.After(2 * time.Second)

	if shippedMetricsCount != 3 {
		t.Errorf("Excepted 3 shipped metrics, got %d", shippedMetricsCount)
	}
}
