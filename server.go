package porthos

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// MethodHandler represents a rpc method handler.
type MethodHandler func(req Request, res Response)

// Server is used to register procedures to be invoked remotely.
type Server interface {
	// Register a method and its handler.
	Register(method string, handler MethodHandler)
	// Register a method, it's handler and it's specification.
	RegisterWithSpec(method string, handler MethodHandler, spec Spec)
	// AddExtension adds extensions to the server instance.
	// Extensions can be used to add custom actions to incoming and outgoing RPC calls.
	AddExtension(ext Extension)
	// ListenAndServe start serving RPC requests.
	ListenAndServe()
	// GetServiceName returns the name of this service.
	GetServiceName() string
	// GetSpecs returns all registered specs.
	GetSpecs() map[string]Spec
	// Close closes the client and AMQP channel.
	// This method returns right after the AMQP channel is closed.
	// In order to give time to the current request to finish (if there's any)
	// it's up to you to wait using the NotifyClose.
	Close()
	// Shutdown shuts down the client and AMQP channel.
	// It provider graceful shutdown, since it will wait the result
	// of <-s.NotifyClose().
	Shutdown()
	// NotifyClose returns a channel to be notified when this server closes.
	NotifyClose() <-chan bool
}

type server struct {
	m              sync.Mutex
	broker         *Broker
	serviceName    string
	channel        *amqp.Channel
	requestChannel <-chan amqp.Delivery
	methods        map[string]MethodHandler
	specs          map[string]Spec
	autoAck        bool
	extensions     []Extension
	topologySet    bool

	closed bool
	closes []chan bool
}

// Options represent all the options supported by the server.
type Options struct {
	AutoAck bool
}

var servePollInterval = 500 * time.Millisecond

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(b *Broker, serviceName string, options Options) (Server, error) {
	s := &server{
		broker:      b,
		serviceName: serviceName,
		methods:     make(map[string]MethodHandler),
		specs:       make(map[string]Spec),
		autoAck:     options.AutoAck,
	}

	err := s.setupTopology()

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *server) setupTopology() error {
	s.m.Lock()
	defer s.m.Unlock()

	var err error
	s.channel, err = s.broker.openChannel()

	if err != nil {
		return err
	}

	// create the response queue
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

	s.topologySet = true

	return nil
}

func (s *server) serve() {
	notifyCh := s.broker.NotifyReestablish()

	for !s.closed {
		if !s.broker.IsConnected() {
			<-notifyCh

			continue
		}

		if !s.topologySet {
			err := s.setupTopology()

			if err != nil {
				log.Printf("[PORTHOS] Error setting up topology after reconnection [%s]", err)
			}

			continue
		}

		s.pipeThroughServerListeningExtensions()
		s.printRegisteredMethods()

		log.Printf("[PORTHOS] Connected to the broker and waiting for incoming rpc requests...")

		for d := range s.requestChannel {
			go func(d amqp.Delivery) {
				err := s.processRequest(d)

				if err != nil {
					log.Printf("[PORTHOS] Error processing request: %s", err)
				}
			}(d)
		}

		s.topologySet = false
	}

	for _, c := range s.closes {
		c <- true
	}
}

func (s *server) printRegisteredMethods() {
	log.Printf("[PORTHOS] [%s]", s.serviceName)

	for method := range s.methods {
		log.Printf("[PORTHOS] . %s", method)
	}
}

func (s *server) processRequest(d amqp.Delivery) error {
	methodName := d.Headers["X-Method"].(string)

	if method, ok := s.methods[methodName]; ok {
		req := &request{s.serviceName, methodName, d.ContentType, d.Body, nil}

		res := newResponse()
		method(req, res)

		ch, err := s.broker.openChannel()

		if err != nil {
			return fmt.Errorf("Error opening channel for response: %s", err)
		}

		if err := ch.Confirm(false); err != nil {
			return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
		}

		confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer ch.Close()

		resWriter := &responseWriter{delivery: d, channel: ch, autoAck: s.autoAck}
		err = resWriter.Write(res)

		if err != nil {
			return fmt.Errorf("Error writing response: %s", err)
		} else {
			if confirmed := <-confirms; !confirmed.Ack {
				return ErrNotAcked
			}
		}
	} else {
		if !s.autoAck {
			d.Reject(false)
		}
		return fmt.Errorf("Method '%s' not found.", methodName)
	}

	return nil
}

func (s *server) pipeThroughServerListeningExtensions() {
	for _, ext := range s.extensions {
		err := ext.ServerListening(s)
		if err != nil {
			log.Printf("[PORTHOS] Error pipe trough server listening. Error %s", err)
		}
	}
}

func (s *server) pipeThroughIncomingExtensions(req Request) {
	for _, ext := range s.extensions {
		ext.IncomingRequest(req)
	}
}

func (s *server) pipeThroughOutgoingExtensions(req Request, res Response, responseTime time.Duration) {
	for _, ext := range s.extensions {
		ext.OutgoingResponse(req, res, responseTime, res.GetStatusCode())
	}
}

func (s *server) Register(method string, handler MethodHandler) {
	s.methods[method] = func(req Request, res Response) {
		s.pipeThroughIncomingExtensions(req)

		started := time.Now()

		// invoke the registered function.
		handler(req, res)

		s.pipeThroughOutgoingExtensions(req, res, time.Since(started))
	}
}

func (s *server) RegisterWithSpec(method string, handler MethodHandler, spec Spec) {
	s.Register(method, handler)
	s.specs[method] = spec
}

// GetServiceName returns the name of this service.
func (s *server) GetServiceName() string {
	return s.serviceName
}

// GetSpecs returns all registered specs.
func (s *server) GetSpecs() map[string]Spec {
	return s.specs
}

func (s *server) AddExtension(ext Extension) {
	s.extensions = append(s.extensions, ext)
}

func (s *server) ListenAndServe() {
	s.serve()
}

func (s *server) Close() {
	s.closed = true
	s.channel.Close()
}

func (s *server) Shutdown() {
	ch := make(chan bool)

	go func() {
		ch <- <-s.NotifyClose()
	}()

	s.Close()
	<-ch
}

func (s *server) NotifyClose() <-chan bool {
	receiver := make(chan bool)
	s.closes = append(s.closes, receiver)

	return receiver
}
