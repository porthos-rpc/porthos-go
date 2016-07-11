package server

import (
    "fmt"
    "encoding/json"

    "github.com/streadway/amqp"
    "github.com/gfronza/porthos/message"
)

type MethodArgs []interface{}
type MethodResponse interface{}

type MethodHandler func(args MethodArgs) MethodResponse

// Server is used to register procedures to be invoked remotely.
type Server struct {
    jobQueue        chan Job
    serviceName     string
    channel         *amqp.Channel
    requestChannel  <-chan amqp.Delivery
    methods         map[string]MethodHandler
}


// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*amqp.Connection, error) {
    return amqp.Dial(amqpURL)
}

// NewServer creates a new instance of Server, responsible for executing remote calls.
func NewServer(conn *amqp.Connection, serviceName string, maxWorkers int) (*Server, error) {
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
        methods:        make(map[string]MethodHandler),
        jobQueue:       make(chan Job, maxWorkers),
    }

    s.startWorkers(maxWorkers)
    s.start()

    return s, nil
}

func (s *Server) startWorkers(maxWorkers int) {
    dispatcher := NewDispatcher(s.jobQueue, maxWorkers)
    dispatcher.Run()
}

func (s *Server) start() {
    go func() {
        for d := range s.requestChannel {
            msg := new(message.MessageBody)
            var err error
            var response []byte
            err = json.Unmarshal(d.Body, msg)

            if err != nil {
                fmt.Println("Unmarshal error: ", err.Error())
                continue
            }

            if method, ok := s.methods[msg.Method]; ok {
                // ack early
                d.Ack(false)

                responseChannel := make(chan MethodResponse)

                // create the job with arguments and a response channel to post the results.
                work := Job{Method: method, Args: msg.Args, ResponseChannel: responseChannel}

                // queue the job.
                s.jobQueue <- work

                // wait for the response asynchronously.
                go func(d amqp.Delivery) {
                    // wait the response
                    res := <-responseChannel

                    close(responseChannel)

                    if res == nil {
                        fmt.Println("Execution without response.")
                        return
                    }

                    response, err = json.Marshal(&res)
                    if err != nil{
                        fmt.Println("Marshal response error: ", err.Error())
                        return
                    }

                    fmt.Printf("Sending response to %s: %s\n", d.CorrelationId, response)

                    err = s.channel.Publish(
                        "",
                        d.ReplyTo,
                        false,
                        false,
                        amqp.Publishing{
                                ContentType:   "application/json",
                                CorrelationId: d.CorrelationId,
                                Body:          response,
                    })

                    if err != nil {
                        fmt.Println("Publish Error: ", err.Error())
                        return
                    }
                }(d)
            } else {
                // TODO: Return timeout?
                fmt.Println("Cannot find method:", msg.Method)
                d.Reject(false)
            }
        }
    }()
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
