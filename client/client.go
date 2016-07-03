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
    "github.com/gfronza/porthos/message"
)

type slot struct {
    InUse bool
    RequestTime time.Time
    ResponseChannel chan interface{}
    TimeoutChannel chan bool
    MessageID uint32
    Index int
}

// Client is an entry point for making remote calls.
type Client struct {
    serviceName string
    defaultTTL int64
    channel *amqp.Channel
    deliveryChannel <-chan amqp.Delivery
    responseQueue *amqp.Queue

    lastMessageID uint32
    slots []slot // client concurrency
    slotsLock *sync.Mutex
}

func (s *slot) getCorrelationID() string {
    return fmt.Sprintf("%d.%d", s.Index, s.MessageID)
}

func (s *slot) free() {
    s.InUse = false
    s.MessageID = 0
    s.Index = -1
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

    slots := make([]slot, 20000)

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
            index, messageID, err := unmarshallCorrelationID(d.CorrelationId)

            if err != nil {
                fmt.Println(err)

                // fail silently.
                continue
            }

            func() {
                c.slotsLock.Lock()
                defer c.slotsLock.Unlock()

                slot := c.slots[index]

                if slot.MessageID == messageID {
                    var jsonResponse interface{}
                    var err error

                    if d.ContentType == "application/json" {
                        err = json.Unmarshal(d.Body, &jsonResponse)
                        if err != nil {
                            fmt.Println("Unmarshal error: ", err.Error())
                        }

                    }

                    if jsonResponse != nil {
                        slot.ResponseChannel <- jsonResponse
                    } else {
                        slot.ResponseChannel <- d.Body
                    }

                    slot.free()
                }
            }()
        }
    }()
}

// Call calls a remote procedure/service with the given name and arguments, using default TTL.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) Call(method string, args ...interface{}) (chan interface{}, chan bool) {
    return c.CallWithTTL(c.defaultTTL, method, args)
}

// CallWithTTL calls a remote procedure/service with the given tll, name and arguments.
// It returns a interface{} channel where you can get you response data from.
func (c *Client) CallWithTTL(ttl int64, method string, args []interface{}) (chan interface{}, chan bool) {
    body, err := json.Marshal(&message.MessageBody{method, args})

    if err != nil {
        panic(err)
    }

    var messageID uint32
    var correlationID string
    var responseChannel chan interface{}
    var timeoutChannel chan bool
    var slot *slot

    func() {
        c.slotsLock.Lock()
        defer c.slotsLock.Unlock()

        slot, err = c.getFreeSlot()

        if err != nil {
            panic(err)
        }

        messageID = slot.MessageID
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
        c.slotsLock.Lock()
        defer c.slotsLock.Unlock()

        // check if this slot is still bing used by the same request
        if slot.MessageID == messageID {
            slot.TimeoutChannel <- true
            slot.free()
        }
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

func (c *Client) getFreeSlot() (*slot, error) {
    for index, s := range c.slots {
        if !s.InUse {
            c.slots[index].InUse = true
            c.slots[index].RequestTime = time.Now()
            c.slots[index].ResponseChannel = make(chan interface{})
            c.slots[index].TimeoutChannel = make(chan bool)
            c.slots[index].MessageID = c.lastMessageID + 1
            c.slots[index].Index = index
            return &c.slots[index], nil
        }
    }

    return nil, errors.New("There's not free slot to get")
}

func unmarshallCorrelationID(correlationID string) (int, uint32, error) {
    s := strings.Split(correlationID, ".")

    if len(s) != 2 {
        return 0, 0, errors.New(fmt.Sprintf("Could not unmarshall correlationID [%s]", correlationID))
    }

    index, err := strconv.Atoi(s[0])

    if err != nil {
        return 0, 0, err
    }

    messageID, err := strconv.ParseUint(s[1], 10, 32)

    if err != nil {
        return 0, 0, err
    }

    return index, uint32(messageID), nil
}
