package client

import (
    "fmt"
    "time"
    "strconv"
    "encoding/json"
    "unsafe"

    "github.com/streadway/amqp"
    "github.com/gfronza/porthos/message"
)

type slot struct {
    InUse bool
    RequestTime time.Time
    ResponseChannel chan interface{}
    TimeoutChannel chan bool
}

// Client is an entry point for making remote calls.
type Client struct {
    serviceName string
    defaultTTL int64
    channel *amqp.Channel
    deliveryChannel <-chan amqp.Delivery
    responseQueue *amqp.Queue
}

func (s *slot) getCorrelationID() string {
    return string(message.UintptrToBytes((uintptr)(unsafe.Pointer(s))))
}

func (s *slot) free() {
    close(s.ResponseChannel)
}

// NewBroker creates a new instance of AMQP connection.
func NewBroker(amqpURL string) (*amqp.Connection, error) {
    return amqp.Dial(amqpURL)
}

// NewClient creates a new instance of Client, responsible for making remote calls.
func NewClient(conn *amqp.Connection, serviceName string, defaultTTL int64) (*Client, error) {
    ch, err := conn.Channel()

    if err != nil {
        conn.Close()
        return nil, err
    }

    // create the response queue (let the amqp server to pick a name for us)
    q, err := ch.QueueDeclare("", false, false, true, false, nil)

    if err != nil {
        ch.Close()
        return nil, err
    }

    dc, err := ch.Consume(
            q.Name, // queue
            "",     // consumer
            true,   // auto-ack
            false,  // exclusive
            false,  // no-local
            false,  // no-wait
            nil,    // args
    )

    if err != nil {
        ch.Close()
        return nil, err
    }

    c := &Client{
        serviceName,
        defaultTTL,
        ch,
        dc,
        &q,
    }

    c.start()

    return c, nil
}

func (c *Client) start() {
    go func() {
        for d := range c.deliveryChannel {
            index := unmarshallCorrelationID(d.CorrelationId)

            func() {
                slot := c.getSlot(index)

                var jsonResponse interface{}
                var err error

                if d.ContentType == "application/json" {
                    err = json.Unmarshal(d.Body, &jsonResponse)
                    if err != nil {
                        fmt.Println("Unmarshal error: ", err.Error())
                    }

                }
                fmt.Printf("Received response %s.\n", d.CorrelationId)
                if jsonResponse != nil {
                    slot.ResponseChannel <- jsonResponse
                } else {
                    slot.ResponseChannel <- d.Body
                }

                slot.free()
            }()
        }
    }()
}

// Call calls a remote procedure/service with the given name and arguments, using default TTL.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) Call(method string, args ...interface{}) (chan interface{}, chan bool) {
    return c.doCallWithTTL(c.defaultTTL, method, args...)
}

// CallWithTTL calls a remote procedure/service with the given tll, name and arguments.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) CallWithTTL(ttl int64, method string, args ...interface{}) (chan interface{}, chan bool) {
    return c.doCallWithTTL(ttl, method, args...)
}

func (c *Client) doCallWithTTL(ttl int64, method string, args ...interface{}) (chan interface{}, chan bool) {
    body, err := json.Marshal(&message.MessageBody{method, args})

    if err != nil {
        panic(err)
    }

    var correlationID string
    var responseChannel chan interface{}
    var timeoutChannel chan bool
    var slot *slot

    func() {
        slot, err = c.getFreeSlot()

        if err != nil {
            panic(err)
        }

        responseChannel = slot.ResponseChannel
        timeoutChannel = slot.TimeoutChannel
        correlationID = slot.getCorrelationID()
    }()

    err = c.channel.Publish(
        "",             // exchange
        c.serviceName,  // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
                Expiration:    strconv.FormatInt(ttl, 10),
                ContentType:   "application/json",
                CorrelationId: correlationID,
                ReplyTo:       c.responseQueue.Name,
                Body:          body,
        })

    if err != nil {
        panic(err)
    }

    // schedule a slot cleanup after TTL value.
    time.AfterFunc(time.Duration(ttl)*time.Millisecond, func() {
        slot.TimeoutChannel <- true
        slot.free()
    })

    return responseChannel, timeoutChannel
}

// CallVoid calls a remote service procedure/service which will not provide any return value.
func (c *Client) CallVoid(method string, args ...interface{}) {
    body, err := json.Marshal(&message.MessageBody{method, args})

    if err != nil {
        panic(err)
    }

    err = c.channel.Publish(
        "",             // exchange
        c.serviceName,  // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
                ContentType:   "application/json",
                Body:          body,
        })

    if err != nil {
        panic(err)
    }
}


// Close the client and AMQP chanel.
func (c *Client) Close() {
    c.channel.Close()
}

func (c *Client) getSlot(index uintptr) *slot {
    return (*slot)(unsafe.Pointer(uintptr(index)))
}

func (c *Client) getFreeSlot()(*slot, error){
    return &slot{
            true,
            time.Now(),
            make(chan interface{}),
            make(chan bool)}, nil
}

func unmarshallCorrelationID(correlationID string) (uintptr) {
    return message.BytesToUintptr([]byte(correlationID))
}
