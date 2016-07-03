package main

import (
    "fmt"
    "os"

    "github.com/gfronza/porthos/client"
)

func main() {
    broker, err := client.NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        fmt.Printf("Error creating broker")
        panic(err)
    }

    defer broker.Close()

    userService, err := client.NewClient(broker, "UserService", 120)

    if err != nil {
        fmt.Printf("Error creating client")
        panic(err)
    }

    defer userService.Close()

    userService.CallVoid("doSomething", 20)
    fmt.Println("Service userService.doSomething invoked")

    var response chan client.Response
    var timeout chan bool

    response, timeout = userService.Call("doSomethingThatReturnsValue", 21)
    fmt.Println("Service userService.doSomethingThatReturnsValue invoked. Waiting for response")

    select {
    case res := <-response:
        fmt.Println("Response: ", res)
    case <-timeout:
        fmt.Println("Timed out :(")
    }

    response, timeout = userService.Call("doSomethingThatReturnsString", "world")
    fmt.Println("Service userService.doSomethingThatReturnsString invoked. Waiting for response")

    select {
    case res := <-response:
        fmt.Println("Response: ", res)
    case <-timeout:
        fmt.Println("Timed out :(")
    }
}
