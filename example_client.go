package main

import (
    "os"
    "fmt"
    "github.com/gfronza/porthos/client"
)

func main() {
    broker, err := client.NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        fmt.Printf("Error creating broker")
        panic(err)
    }

    userService, err := client.NewClient(broker, "UserService")

    if err != nil {
        fmt.Printf("Error creating client")
        panic(err)
    }

    userService.CallVoid(120, "doSomething", 20)
    fmt.Println("Service userService.doSomething executed")

    ch := userService.Call(50, "doSomethingThatReturnsValue", 20)
    fmt.Println("Service userService.doSomethingThatReturnsValue executed. Waiting for response")

    // consume the response
    response := <-ch

    if response.Timeout {
        fmt.Println("Timedout")
    } else {
        fmt.Printf("Response: %#v\n", response.Data)
    }

    userService.Close()
}
