package server

import "github.com/streadway/amqp"

type methodHandler func(args ...interface{}) interface{}

// Server is used to register procedures to be invoked remotely.
type Server struct {
	serviceName    string
	channel        *amqp.Channel
	requestChannel <-chan amqp.Delivery
	methods        map[string]methodHandler
}

// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*amqp.Connection, error) {
	return amqp.Dial(amqpURL)
}

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(conn *amqp.Connection, serviceName string) (*Server, error) {
	ch, err := conn.Channel()

	if err != nil {
		conn.Close()
		return nil, err
	}

	// create the response queue (let the amqp server to pick a name for us)
	_, err = ch.QueueDeclare(
		serviceName, // name
		true,        // durable
		false,       // delete when usused
		true,        // exclusive
		false,       // noWait
		nil,         // arguments
	)

	if err != nil {
		ch.Close()
		return nil, err
	}

	dc, err := ch.Consume(
		serviceName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)

	if err != nil {
		ch.Close()
		return nil, err
	}

	s := &Server{
		serviceName:    serviceName,
		channel:        ch,
		requestChannel: dc,
		methods:        make(map[string]methodHandler),
	}

	s.start()

	return s, nil
}

func (s *Server) start() {
	go func() {
		for d := range s.requestChannel {
			// Ack early
			d.Ack(false)

			// TODO: unmarshall body and call the method.
			// we need to decide whether we call the method in the same context
			// or inside a goroutine. We could we a "worker pool" as well,
			// to limit concurrent executions
		}
	}()
}

// Register a method and its handler.
func (s *Server) Register(method string, handler methodHandler) {
	s.methods[method] = handler
}

// Close the client and AMQP chanel.
func (s *Server) Close() {
	s.channel.Close()
}
