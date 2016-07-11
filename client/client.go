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

type Slot struct {
    requestTime time.Time
    ResponseChannel chan interface{}
//    TimeoutChannel chan bool
}

// Client is an entry point for making remote calls.
type Client struct {
    serviceName string
    defaultTTL int64
    channel *amqp.Channel
    deliveryChannel <-chan amqp.Delivery
    responseQueue *amqp.Queue
}

func (s *Slot) getCorrelationID() string {
    return string(message.UintptrToBytes((uintptr)(unsafe.Pointer(s))))
}

func (s *Slot) free() {
    close(s.ResponseChannel)
    //s.ResponseChannel = nil
    //fmt.Println("################################################ close ############################################ ", s.ResponseChannel)
}

func (s *Slot) GetRequestTime() time.Time {
    return s.requestTime
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
            address := unmarshallCorrelationID(d.CorrelationId)

            func() {
                slot := c.getSlot(address)

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
                //slot.free()
                
            }()
        }
    }()
}

// Call calls a remote procedure/service with the given name and arguments, using default TTL.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) Call(method string, args ...interface{}) (*Slot) {
    return c.doCallWithTTL(c.defaultTTL, method, args...)
}

// CallWithTTL calls a remote procedure/service with the given tll, name and arguments.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) CallWithTTL(ttl int64, method string, args ...interface{}) (*Slot) {
    return c.doCallWithTTL(ttl, method, args...)
}

func (c *Client) doCallWithTTL(ttl int64, method string, args ...interface{}) (*Slot) {
    body, err := json.Marshal(&message.MessageBody{method, args})

    if err != nil {
        panic(err)
    }

    slot := c.getFreeSlot()

    err = c.channel.Publish(
        "",             // exchange
        c.serviceName,  // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
                Expiration:    strconv.FormatInt(ttl, 10),
                ContentType:   "application/json",
                CorrelationId: slot.getCorrelationID(),
                ReplyTo:       c.responseQueue.Name,
                Body:          body,
        })

    if err != nil {
        panic(err)
    }

    // schedule a slot cleanup after TTL value.
    //time.AfterFunc(time.Duration(ttl)*time.Millisecond, func() {      slot.TimeoutChannel <- true   })

    return slot
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

func (c *Client) getSlot(address uintptr) *Slot {
    return (*Slot)(unsafe.Pointer(uintptr(address)))
}

func (c *Client) getFreeSlot()(*Slot){
    return &Slot{
            time.Now(),
            make(chan interface{}),
            //(chan bool, 1),
        }
}

func unmarshallCorrelationID(correlationID string) (uintptr) {
    return message.BytesToUintptr([]byte(correlationID))
}
