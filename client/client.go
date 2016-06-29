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

type Response struct {
    Data interface{}
    Timeout bool
}

type Slot struct {
    InUse bool
    RequestTime time.Time
    ResponseChannel chan Response
    MessageId uint32
    Index int
}

type Client struct {
    serviceName string
    connection *amqp.Connection
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

func NewBroker(amqpUrl string) (*amqp.Connection, error) {
    return amqp.Dial(amqpUrl)
}

func NewClient(conn *amqp.Connection, serviceName string) (*Client, error) {
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
        conn,
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

            c.slotsLock.Lock()
            slot := c.slots[index]

            if slot.MessageId == messageId {
                var jsonResponse interface{}

                if d.ContentType == "application/json" && json.Unmarshal(d.Body, jsonResponse) == nil {
                    slot.ResponseChannel <- Response{jsonResponse, false}
                } else {
                    slot.ResponseChannel <- Response{d.Body, false}
                }

                slot.Free()
            }
            c.slotsLock.Unlock()
        }
    }()
}

func (c *Client) Call(ttl int64, method string, args ...interface{}) chan Response {
    arguments, err := json.Marshal(args)

    if err != nil {
        panic(err)
    }

    var messageId uint32
    var correlationId string
    var responseChannel chan Response
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
            slot.ResponseChannel <- Response{nil, true}
            slot.Free()
        }
    })

    return responseChannel
}

// Call a remote service method with will not provide any return value.
func (c *Client) CallVoid(ttl int64, method string, args ...interface{}) {
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
                Expiration:    strconv.FormatInt(ttl, 10),
                ContentType:   "application/json",
                Body:          []byte(arguments),
        })

    if err != nil {
        panic(err)
    }
}

func (c *Client) Close() {
    c.connection.Close()
    c.channel.Close()
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

func (c *Client) getFreeSlot() (*Slot, error) {
    for index, s := range c.slots {
        if !s.InUse {
            c.slots[index].InUse = true
            c.slots[index].RequestTime = time.Now()
            c.slots[index].ResponseChannel = make(chan Response)
            c.slots[index].MessageId = c.lastMessageId + 1
            c.slots[index].Index = index
            return &c.slots[index], nil
        }
    }

    return nil, errors.New("There's not free slot to get")
}
