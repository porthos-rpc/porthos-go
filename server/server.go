package server

import (
	"time"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

// MethodHandler represents a rpc method handler.
type MethodHandler func(req Request, res Response)

// Server is used to register procedures to be invoked remotely.
type Server struct {
	jobQueue       chan Job
	serviceName    string
	channel        *amqp.Channel
	requestChannel <-chan amqp.Delivery
	methods        map[string]MethodHandler
	autoAck        bool
	extensions     []*Extension
	maxWorkers     int
}

// Options represent all the options supported by the server.
type Options struct {
	MaxWorkers int
	AutoAck    bool
}

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(b *broker.Broker, serviceName string, options Options) (*Server, error) {
	ch, err := b.Conn.Channel()

	if err != nil {
		b.Conn.Close()
		return nil, err
	}

	// create the response queue (let the amqp server to pick a name for us)
	_, err = ch.QueueDeclare(
		serviceName, // name
		true,        // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // noWait
		nil,         // arguments
	)

	if err != nil {
		ch.Close()
		return nil, err
	}

	dc, err := ch.Consume(
		serviceName,     // queue
		"",              // consumer
		options.AutoAck, // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)

	if err != nil {
		ch.Close()
		return nil, err
	}

	maxWorkers := options.MaxWorkers

	if maxWorkers <= 0 {
		maxWorkers = 100
	}

	s := &Server{
		serviceName:    serviceName,
		channel:        ch,
		requestChannel: dc,
		methods:        make(map[string]MethodHandler),
		jobQueue:       make(chan Job, maxWorkers),
		autoAck:        options.AutoAck,
		maxWorkers:     maxWorkers,
	}

	return s, nil
}

func (s *Server) startWorkers(maxWorkers int) {
	dispatcher := NewDispatcher(s.jobQueue, maxWorkers)
	dispatcher.Run()
}

func (s *Server) start() {
	go func() {
		for d := range s.requestChannel {
			s.processRequest(d)
		}
	}()
}

func (s *Server) processRequest(d amqp.Delivery) {
	methodName := d.Headers["X-Method"].(string)

	if method, ok := s.methods[methodName]; ok {
		req := &request{s.serviceName, methodName, d.ContentType, d.Body}

		resWriter := &responseWriter{delivery: d, channel: s.channel, autoAck: s.autoAck}

		// create the job with arguments and a response writer.
		work := Job{Method: method, Request: req, ResponseWriter: resWriter}

		// queue the job.
		s.jobQueue <- work
	} else {
		log.Error("Method '%s' not found.", methodName)
		if !s.autoAck {
			d.Reject(false)
		}
	}
}

func (s *Server) pipeThroughIncomingExtensions(req Request) {
	for _, ext := range s.extensions {
		if ext.incoming != nil {
			ext.incoming <- IncomingRPC{req}
		}
	}
}

func (s *Server) pipeThroughOutgoingExtensions(req Request, res Response, responseTime time.Duration) {
	for _, ext := range s.extensions {
		if ext.outgoing != nil {
			ext.outgoing <- OutgoingRPC{req, res, responseTime, res.GetStatusCode()}
		}
	}
}

// Register a method and its handler.
func (s *Server) Register(method string, handler MethodHandler) {
	s.methods[method] = func(req Request, res Response) {
		s.pipeThroughIncomingExtensions(req)

		started := time.Now()

		// invoke the registered function.
		handler(req, res)

		s.pipeThroughOutgoingExtensions(req, res, time.Since(started))
	}
}

// AddExtension adds extensions to the server instance.
// Extensions can be used to add custom actions to incoming and outgoing RPC calls.
func (s *Server) AddExtension(ext *Extension) {
	s.extensions = append(s.extensions, ext)
}

// Start serving RPC requests.
func (s *Server) Start() {
	s.startWorkers(s.maxWorkers)
	s.start()
}

// Close the client and AMQP chanel.
func (s *Server) Close() {
	s.channel.Close()
}
