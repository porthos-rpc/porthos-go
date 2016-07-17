package server

import (
    "fmt"
    "encoding/json"

    "github.com/streadway/amqp"
    "github.com/gfronza/porthos/message"
)

type Request struct {
    args []interface{}
}

type Response struct {
    content     []byte
    contentType string
}

type MethodHandler func(req Request, res *Response)

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
        false,       // exclusive
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

// GetArg returns an argument giving the index.
func (r *Request) GetArg(index int) *Argument {
    return &Argument{r.args[index]}
}

// SetContent sets the content of the method's response.
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

// GetEncodedContent returns the method's response encoded in JSON format.
func (r *Response) GetContent() []byte {
    return r.content
}

// GetEncodedContent returns the method's response encoded in JSON format.
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
        fmt.Println("Unmarshal error: ", err.Error())
        return
    }

    if method, ok := s.methods[msg.Method]; ok {
        // ack early
        d.Ack(false)

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
                fmt.Println("Error encoding response content: ", err.Error())
                return
            }

            fmt.Println("Sending response: ", resContent)

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
