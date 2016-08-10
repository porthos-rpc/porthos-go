package server

import (
	"encoding/json"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/message"
	"github.com/streadway/amqp"
)

// Request represents a rpc request.
type Request struct {
	args []interface{}
}

// Response represents a rpc response.
type Response struct {
	content     []byte
	contentType string
}

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
}

// ServerOptions represent all the options supported by the server.
type ServerOptions struct {
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
func NewServer(broker *Broker, serviceName string, options ServerOptions) (*Server, error) {
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

// GetArg returns an argument giving the index.
func (r *Request) GetArg(index int) *Argument {
	return &Argument{r.args[index]}
}

// JSON sets the content of the method's response.
func (r *Response) JSON(c interface{}) {
	if c == nil {
		panic("Response content is empty")
	}

	jsonContent, err := json.Marshal(&c)

	if err != nil {
		panic(err)
	}

	r.content = jsonContent
	r.contentType = "application/json"
}

// GetContent returns the method's response encoded in JSON format.
func (r *Response) GetContent() []byte {
	return r.content
}

// GetContentType returns the method's response encoded in JSON format.
func (r *Response) GetContentType() string {
	return r.contentType
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
		responseChannel := make(chan *Response)

		req := Request{msg.Args}

		// create the job with arguments and a response channel to post the results.
		work := Job{Method: method, Request: req, ResponseChannel: responseChannel}

		// queue the job.
		s.jobQueue <- work

		// wait for the response asynchronously.
		go func(d amqp.Delivery) {
			// wait the response
			res := <-responseChannel

			close(responseChannel)

			resContent := res.GetContent()
			resContentType := res.GetContentType()

			if err != nil {
				log.Error("Error encoding response content: '%s'", err.Error())
				return
			}

			log.Info("Sending response to queue '%s'. Slot: '%d'", d.ReplyTo, []byte(d.CorrelationId))

			err = s.channel.Publish(
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   resContentType,
					CorrelationId: d.CorrelationId,
					Body:          resContent,
				})

			if err != nil {
				log.Error("Publish Error: '%s'", err.Error())
				return
			}

			if !s.autoAck {
				d.Ack(false)
				log.Debug("Ack method '%s' from slot '%d'", msg.Method, []byte(d.CorrelationId))
			}
		}(d)
	} else {
		log.Error("Method '%s' not found.", msg.Method)
		if !s.autoAck {
			d.Reject(false)
		}
	}
}

// Register a method and its handler.
func (s *Server) Register(method string, handler MethodHandler) {
	s.methods[method] = handler
}

// Close the client and AMQP chanel.
func (s *Server) Close() {
	s.channel.Close()
}

// ServeForever blocks the current context to serve remote requests forever.
func (s *Server) ServeForever() {
	select {}
}
