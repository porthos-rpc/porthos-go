package client

import (
    "fmt"
    "time"
    "sync"
    "errors"
    "strings"
    "strconv"
    "encoding/json"

    "github.com/streadway/amqp"
)

type Response interface{}

type Slot struct {
    InUse bool
    RequestTime time.Time
    ResponseChannel chan Response
    TimeoutChannel chan bool
    MessageId uint32
    Index int
}

type Client struct {
    serviceName string
    defaultTTL int64
    channel *amqp.Channel
    deliveryChannel <-chan amqp.Delivery
    responseQueue *amqp.Queue

    lastMessageId uint32
    slots []Slot // client concurrency
    slotsLock *sync.Mutex
}

func (s *Slot) GetCorrelationId() string {
    return fmt.Sprintf("%d.%d", s.Index, s.MessageId)
}

func (s *Slot) Free() {
    s.InUse = false
    s.MessageId = 0
    s.Index = -1
    close(s.ResponseChannel)
}

// Creates a new instance of AMQP connection.
func NewBroker(amqpUrl string) (*amqp.Connection, error) {
    return amqp.Dial(amqpUrl)
}

// Creates a new instance of Client, responsible for making remote calls.
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

    slots := make([]Slot, 20000)

    c := &Client{
        serviceName,
        defaultTTL,
        ch,
        dc,
        &q,
        0,
        slots,
        &sync.Mutex{},
    }

    c.start()

    return c, nil
}

func (c *Client) start() {
    go func() {
        for d := range c.deliveryChannel {
            index, messageId, err := unmarshallCorrelationId(d.CorrelationId)

            if err != nil {
                fmt.Println(err)

                // fail silently.
                continue
            }

            func() {
                c.slotsLock.Lock()
                defer c.slotsLock.Unlock()

                slot := c.slots[index]

                if slot.MessageId == messageId {
                    var jsonResponse interface{}

                    if d.ContentType == "application/json" && json.Unmarshal(d.Body, jsonResponse) == nil {
                        slot.ResponseChannel <- jsonResponse
                    } else {
                        slot.ResponseChannel <- d.Body
                    }

                    slot.Free()
                }
            }()
        }
    }()
}

// Calls a remote procedure/service with the given name and arguments, using default TTL.
// It returns a Response channel where you can get you response data from.
func (c *Client) Call(method string, args ...interface{}) (chan Response, chan bool) {
    return c.CallWithTTL(c.defaultTTL, method, args)
}

// Calls a remote procedure/service with the given tll, name and arguments.
// It returns a Response channel where you can get you response data from.
func (c *Client) CallWithTTL(ttl int64, method string, args ...interface{}) (chan Response, chan bool) {
    arguments, err := json.Marshal(args)

    if err != nil {
        panic(err)
    }

    var messageId uint32
    var correlationId string
    var responseChannel chan Response
    var timeoutChannel chan bool
    var slot *Slot

    func() {
        c.slotsLock.Lock()
        defer c.slotsLock.Unlock()

        slot, err = c.getFreeSlot()

        if err != nil {
            panic(err)
        }

        messageId = slot.MessageId
        responseChannel = slot.ResponseChannel
        timeoutChannel = slot.TimeoutChannel
        correlationId = slot.GetCorrelationId()
    }()

    err = c.channel.Publish(
        "",             // exchange
        c.serviceName,  // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
                Expiration:    strconv.FormatInt(ttl, 10),
                ContentType:   "application/json",
                CorrelationId: correlationId,
                ReplyTo:       c.responseQueue.Name,
                Body:          []byte(arguments),
        })

    if err != nil {
        panic(err)
    }

    // schedule a slot cleanup after TTL value.
    time.AfterFunc(time.Duration(ttl)*time.Millisecond, func() {
        c.slotsLock.Lock()
        defer c.slotsLock.Unlock()

        // check if this slot is still bing used by the same request
        if slot.MessageId == messageId {
            slot.TimeoutChannel <- true
            slot.Free()
        }
    })

    return responseChannel, timeoutChannel
}

// Call a remote service procedure/service which will not provide any return value.
func (c *Client) CallVoid(method string, args ...interface{}) {
    arguments, err := json.Marshal(args)

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
                Body:          []byte(arguments),
        })

    if err != nil {
        panic(err)
    }
}


// Close the client and AMQP chanel.
func (c *Client) Close() {
    c.channel.Close()
}

func (c *Client) getFreeSlot() (*Slot, error) {
    for index, s := range c.slots {
        if !s.InUse {
            c.slots[index].InUse = true
            c.slots[index].RequestTime = time.Now()
            c.slots[index].ResponseChannel = make(chan Response)
            c.slots[index].TimeoutChannel = make(chan bool)
            c.slots[index].MessageId = c.lastMessageId + 1
            c.slots[index].Index = index
            return &c.slots[index], nil
        }
    }

    return nil, errors.New("There's not free slot to get")
}

func unmarshallCorrelationId(correlationId string) (int, uint32, error) {
    s := strings.Split(correlationId, ".")

    if len(s) != 2 {
        return 0, 0, errors.New(fmt.Sprintf("Could not unmarshall correlationId [%s]", correlationId))
    }

    index, err := strconv.Atoi(s[0])

    if err != nil {
        return 0, 0, err
    }

    messageId, err := strconv.ParseUint(s[1], 10, 32)

    if err != nil {
        return 0, 0, err
    }

    return index, uint32(messageId), nil
}
