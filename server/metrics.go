package server

import (
	"encoding/json"
	"time"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

const metricsQueueName = "porthos.metrics"

// MetricsShipperConfig defines config params for the NewMetricsShipperExtension.
type MetricsShipperConfig struct {
	BufferSize int
}

type metricEntry struct {
	MethodName   string        `json:"methodName"`
	ResponseTime time.Duration `json:"responsetime"`
}

type metricsCollector struct {
	channel *amqp.Channel
	index   int
	buffer  []*metricEntry
}

func (mc *metricsCollector) append(entry *metricEntry) {
	mc.buffer[mc.index] = entry

	mc.index++
}

func (mc *metricsCollector) reset() {
	for i := range mc.buffer {
		mc.buffer[i] = nil
	}

	mc.index = 0
}

func (mc *metricsCollector) isFull() bool {
	return mc.index == len(mc.buffer)
}

func (mc *metricsCollector) ship() {
	log.Debug("Shipping metrics to broker...")

	payload, err := json.Marshal(mc.buffer)

	if err != nil {
		log.Error("Error json encoding metrics payload", err)
		return
	}

	err = mc.channel.Publish(
		"",
		metricsQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})

	if err != nil {
		log.Error("Error publishing metrics to the broker", err)
	}
}

// NewMetricsShipperExtension creates a new extension that collect stats from RPC calls (request and response)
func NewMetricsShipperExtension(broker *Broker, config MetricsShipperConfig) *Extension {
	ext := NewOutgoingOnlyExtension()

	ch, err := broker.conn.Channel()

	if err != nil {
		log.Error("Error creating metrics broker channel", err)
		return nil
	}

	_, err = ch.QueueDeclare(
		metricsQueueName, // name
		true,             // durable
		false,            // delete when usused
		false,            // exclusive
		false,            // noWait
		nil,              // arguments
	)

	mc := metricsCollector{
		channel: ch,
		buffer:  make([]*metricEntry, config.BufferSize, config.BufferSize),
	}

	go func() {
		log.Info("Metrics shipper extension is waiting for outgoing events...")

		for {
			out := <-ext.Outgoing()

			mc.append(&metricEntry{out.Request.MethodName, out.ResponseTime})

			if mc.isFull() {
				mc.ship()
				mc.reset()
			}
		}
	}()

	return ext
}
