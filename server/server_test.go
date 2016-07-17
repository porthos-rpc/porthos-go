package server

import (
    "os"
    "time"
    "testing"
    "encoding/json"

    "github.com/streadway/amqp"
    "github.com/gfronza/porthos/message"
)

type ResponseTest struct {
    Original    float64 `json:"original"`
    Sum         float64 `json:"sum"`
}

func TestNewServer(t *testing.T) {
    broker, err := NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        t.Fatal("NewBroker failed.", err)
    }

    userService, err := NewServer(broker, "UserService", 1)
    defer userService.Close()

    if err != nil {
        t.Fatal("NewServer failed.", err)
    }

    if userService.jobQueue == nil {
        t.Error("Service jobQueue is nil.")
    }

    if userService.serviceName != "UserService" {
        t.Errorf("Wrong serviceName, expected: 'UserService', got: '%s'", userService.serviceName)
    }

    if userService.channel == nil {
        t.Error("Service channel is nil.")
    }

    if userService.requestChannel == nil {
        t.Error("Service requestChannel is nil.")
    }

    if userService.methods == nil {
        t.Error("Service methods is nil.")
    }
}

func TestServerProcessRequest(t *testing.T) {
    broker, err := NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        t.Fatal("NewBroker failed.", err)
    }

    userService, err := NewServer(broker, "UserService", 1)
    defer userService.Close()

    if err != nil {
        t.Fatal("NewServer failed.", err)
    }

    // register the method that we will test.
    userService.Register("doSomething", func(req Request, res *Response) {
        x := req.GetArg(0).AsFloat64()

        res.JSON(ResponseTest{x, x+1})
    })

    // This code below is to simulate the client invoking the remote method.

    // create the request message body.
    body, _ := json.Marshal(&message.MessageBody{"doSomething", []interface{}{10}})

    // declare the response queue.
    q, err := userService.channel.QueueDeclare("", false, false, true, false, nil)

    if err != nil {
        t.Fatal("Queue declare failed.", err)
    }

    // start consuming from the response queue.
    dc, err := userService.channel.Consume(
        q.Name, // queue
        "",                // consumer
        true,              // auto-ack
        false,             // exclusive
        false,             // no-local
        false,             // no-wait
        nil,               // args
    )

    if err != nil {
        t.Fatal("Consume failed.", err)
    }

    // publish the request.
    err = userService.channel.Publish(
        "",
        userService.serviceName,
        false,
        false,
        amqp.Publishing{
                Expiration:    "3000",
                ContentType:   "application/json",
                CorrelationId: "1",
                ReplyTo:       q.Name,
                Body:          body,
        })

        if err != nil {
            t.Fatal("Publish failed.", err)
        }

    // wait the response or timeout.
    for {
        select {
            case response := <-dc:
                var responseTest ResponseTest
                err = json.Unmarshal(response.Body, &responseTest)

                if responseTest.Sum != 11 {
                    t.Errorf("Response failed, expected: 11, got: %s", responseTest.Sum)
                }

                return
            case <-time.After(5 * time.Second):
                t.Fatal("No response receive. Timedout.")
        }
    }
}
