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

    // create a client with a default timeout of 1 second.
    userService, err := client.NewClient(broker, "UserService", 1000)

    if err != nil {
        fmt.Printf("Error creating client")
        panic(err)
    }

    defer userService.Close()

    // call a remote method that is "void".
    userService.CallVoid("doSomething", 20)
    fmt.Println("Service userService.doSomething invoked")

    // call a lot of methods concurrently
    for i := 0; i < 100000; i++ {
        go func(idx int) {
            slot := userService.Call("doSomethingThatReturnsValue", idx)
            fmt.Printf("Service userService.doSomethingThatReturnsValue invoked %d\n", idx)

            select {
            case res := <-slot.ResponseChannel:
                data := res.(map[string]interface{})
                fmt.Printf("Response %d. Original: %f. Sum: %f\n", idx, data["original"], data["sum"])
            case <-slot.TimeoutChannel:
                fmt.Printf("Timed out %d :(\n", idx)
            }
        }(i)
	}

    go func(){
        // call a method with a custom timeout.
        slot := userService.CallWithTTL(200, "doSomethingThatReturnsValue", 21)
        fmt.Println("Service userService.doSomethingThatReturnsValue invoked. Waiting for response")

        select {
        case res := <-slot.ResponseChannel:
            fmt.Println("Response1: ", res)
        case <-slot.TimeoutChannel:
            fmt.Println("Timed out :(")
        }
    }()
    // wait forever (to give time to execute all goroutines)
    select{}
}
