package porthos

import (
	"encoding/json"
	"time"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

const metricsQueueName = "porthos.metrics"

type metricEntry struct {
	ServiceName  string        `json:"serviceName"`
	MethodName   string        `json:"methodName"`
	ResponseTime time.Duration `json:"responsetime"`
	StatusCode   int16         `json:"statusCode"`
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

// MetricsShipperConfig defines config params for the NewMetricsShipperExtension.
type MetricsShipperConfig struct {
	BufferSize int
}

// MetricsShipperExtension logs incoming requests and outgoing responses.
type MetricsShipperExtension struct {
	mc *metricsCollector
}

// ServerListening this is not implemented in this extension.
func (a *MetricsShipperExtension) ServerListening(server Server) {}

// IncomingRequest this is not implemented in this extension.
func (a *MetricsShipperExtension) IncomingRequest(req Request) {}

// OutgoingResponse ships metrics based on responses to the broker.
func (a *MetricsShipperExtension) OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int16) {
	a.mc.append(&metricEntry{req.GetServiceName(), req.GetMethodName(), resTime, statusCode})

	if a.mc.isFull() {
		a.mc.ship()
		a.mc.reset()
	}
}

// NewMetricsShipperExtension creates a new extension that logs everything to stdout.
func NewMetricsShipperExtension(b *Broker, config MetricsShipperConfig) Extension {
	ch, err := b.openChannel()

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

	log.Info("Metrics shipper extension is waiting for outgoing events...")

	return &MetricsShipperExtension{&metricsCollector{
		channel: ch,
		buffer:  make([]*metricEntry, config.BufferSize, config.BufferSize),
	}}
}
