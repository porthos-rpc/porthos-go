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

    response1, timeout1 := userService.CallWithTTL(200, "doSomethingThatReturnsValue", 21)
    fmt.Println("Service userService.doSomethingThatReturnsValue invoked. Waiting for response")

    response2, timeout2 := userService.Call("doSomethingThatReturnsString", "world")
    fmt.Println("Service userService.doSomethingThatReturnsString invoked. Waiting for response")

    select {
    case res := <-response1:
        fmt.Println("Response: ", res)
    case <-timeout1:
        fmt.Println("Timed out :(")
    }

    select {
    case res := <-response2:
        fmt.Println("Response: ", res)
    case <-timeout2:
        fmt.Println("Timed out :(")
    }
}
