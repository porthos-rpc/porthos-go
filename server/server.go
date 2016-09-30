package server

import (
	"encoding/json"
	"time"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/message"
	"github.com/streadway/amqp"
)

// MethodHandler represents a rpc method handler.
type MethodHandler func(req Request, res *Response)

// Broker holds an implementation-specific connection.
type Broker struct {
	conn *amqp.Connection
}

// Server is used to register procedures to be invoked remotely.
type Server struct {
	jobQueue       chan Job
	serviceName    string
	channel        *amqp.Channel
	requestChannel <-chan amqp.Delivery
	methods        map[string]MethodHandler
	autoAck        bool
	extensions     []*Extension
}

// Options represent all the options supported by the server.
type Options struct {
	MaxWorkers int
	AutoAck    bool
}

// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*Broker, error) {
	conn, err := amqp.Dial(amqpURL)

	if err != nil {
		return nil, err
	}

	return &Broker{conn}, nil
}

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(broker *Broker, serviceName string, options Options) (*Server, error) {
	ch, err := broker.conn.Channel()

	if err != nil {
		broker.conn.Close()
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
	}

	s.startWorkers(maxWorkers)
	s.start()

	return s, nil
}

// Close the broker connection.
func (b *Broker) Close() {
	b.conn.Close()
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
	msg := new(message.MessageBody)

	err := json.Unmarshal(d.Body, msg)

	if err != nil {
		log.Error("Unmarshal error: %s", err.Error())
		return
	}

	if method, ok := s.methods[msg.Method]; ok {
		req := Request{s.serviceName, msg.Method, msg.Args, d.Body}

		resWriter := ResponseWriter{delivery: d, channel: s.channel, autoAck: s.autoAck}

		// create the job with arguments and a response writer.
		work := Job{Method: method, Request: req, ResponseWriter: resWriter}

		// queue the job.
		s.jobQueue <- work
	} else {
		log.Error("Method '%s' not found.", msg.Method)
		if !s.autoAck {
			d.Reject(false)
		}
	}
}

func (s *Server) pipeThroughIncomingExtensions(req *Request) {
	for _, ext := range s.extensions {
		if ext.incoming != nil {
			ext.incoming <- IncomingRPC{req}
		}
	}
}

func (s *Server) pipeThroughOutgoingExtensions(req *Request, res *Response, responseTime time.Duration) {
	for _, ext := range s.extensions {
		if ext.outgoing != nil {
			ext.outgoing <- OutgoingRPC{req, res, responseTime, res.statusCode}
		}
	}
}

// Register a method and its handler.
func (s *Server) Register(method string, handler MethodHandler) {
	s.methods[method] = func(req Request, res *Response) {
		s.pipeThroughIncomingExtensions(&req)

		started := time.Now()

		// invoke the registered function.
		handler(req, res)

		s.pipeThroughOutgoingExtensions(&req, res, time.Since(started))
	}
}

// AddExtension adds extensions to the server instance.
// Extensions can be used to add custom actions to incoming and outgoing RPC calls.
func (s *Server) AddExtension(ext *Extension) {
	s.extensions = append(s.extensions, ext)
}

// Close the client and AMQP chanel.
func (s *Server) Close() {
	s.channel.Close()
}

// ServeForever blocks the current context to serve remote requests forever.
func (s *Server) ServeForever() {
	select {}
}
