package server

import (
	"sync"
	"time"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/streadway/amqp"
)

// MethodHandler represents a rpc method handler.
type MethodHandler func(req Request, res Response)

// Server is used to register procedures to be invoked remotely.
type Server struct {
	m              sync.Mutex
	broker         *broker.Broker
	jobQueue       chan Job
	serviceName    string
	channel        *amqp.Channel
	requestChannel <-chan amqp.Delivery
	methods        map[string]MethodHandler
	autoAck        bool
	extensions     []*Extension
	maxWorkers     int
	serving        bool

	closed bool
	closes []chan bool
}

// Options represent all the options supported by the server.
type Options struct {
	MaxWorkers int
	AutoAck    bool
}

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(b *broker.Broker, serviceName string, options Options) (*Server, error) {
	maxWorkers := options.MaxWorkers

	if maxWorkers <= 0 {
		maxWorkers = 100
	}

	s := &Server{
		broker:      b,
		serviceName: serviceName,
		methods:     make(map[string]MethodHandler),
		jobQueue:    make(chan Job, maxWorkers),
		autoAck:     options.AutoAck,
		maxWorkers:  maxWorkers,
	}

	err := s.setupTopology()

	if err != nil {
		return nil, err
	}

	go s.handleReestablishedConnnection()

	return s, nil
}

func (s *Server) setupTopology() error {
	s.m.Lock()
	defer s.m.Unlock()

	var err error
	s.channel, err = s.broker.OpenChannel()

	if err != nil {
		return err
	}

	// create the response queue (let the amqp server to pick a name for us)
	_, err = s.channel.QueueDeclare(
		s.serviceName, // name
		true,          // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // noWait
		nil,           // arguments
	)

	if err != nil {
		s.channel.Close()
		return err
	}

	s.requestChannel, err = s.channel.Consume(
		s.serviceName, // queue
		"",            // consumer
		s.autoAck,     // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)

	if err != nil {
		s.channel.Close()
		return err
	}

	return nil
}

func (s *Server) handleReestablishedConnnection() {
	for !s.closed {
		<-s.broker.NotifyReestablish()

		err := s.setupTopology()

		if err == nil {
			if s.serving {
				go s.serve()
			}
		} else {
			log.Error("Error setting up topology after reconnection", err)
		}
	}
}

func (s *Server) startWorkers(maxWorkers int) {
	dispatcher := NewDispatcher(s.jobQueue, maxWorkers)
	dispatcher.Run()
}

func (s *Server) serve() {
	s.printRegisteredMethods()

	log.Info("Connected to the broker and waiting for incoming rpc requests...")

	for d := range s.requestChannel {
		s.processRequest(d)
	}

	if s.closed {
		for _, c := range s.closes {
			c <- true
		}
	}
}

func (s *Server) printRegisteredMethods() {
	log.Info("[%s]", s.serviceName)

	for method := range s.methods {
		log.Info(". %s", method)
	}
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

// ListenAndServe start serving RPC requests.
func (s *Server) ListenAndServe() {
	s.serving = true

	s.startWorkers(s.maxWorkers)
	go s.serve()
}

// Close the client and AMQP channel.
// This method returns right after the AMQP channel is closed.
// In order to give time to the current request to finish (if there's one)
// it's up to you to wait using the NotifyClose.
func (s *Server) Close() {
	s.closed = true
	s.channel.Close()
}

// Shutdown shuts down the client and AMQP channel.
// It provider graceful shutdown, since it will wait the result
// of <-s.NotifyClose().
func (s *Server) Shutdown() {
	ch := make(chan bool)

	go func() {
		ch <- <-s.NotifyClose()
	}()

	s.Close()
	<-ch
}

// NotifyClose returns a channel to be notified then this server closes.
func (s *Server) NotifyClose() <-chan bool {
	receiver := make(chan bool)
	s.closes = append(s.closes, receiver)

	return receiver
}
