package main

import (
    "fmt"
    "os"

    "github.com/gfronza/porthos/server"
)

func main() {
    broker, err := server.NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        fmt.Printf("Error creating broker")
        panic(err)
    }

    defer broker.Close()

    userService, err := server.NewServer(broker, "UserService", 150)

    if err != nil {
        fmt.Printf("Error creating server")
        panic(err)
    }

    defer userService.Close()

    userService.Register("doSomething", func(args server.MethodArgs) server.MethodResponse {
        return nil
    })

    userService.Register("doSomethingThatReturnsValue", func(args server.MethodArgs) server.MethodResponse {
        type test struct {
            Original    float64 `json:"original"`
            Sum         float64 `json:"sum"`
        }

        x := args[0].(float64)

        return test{x, x+1}
    })

    userService.Register("doSomethingThatReturnsString", func(args server.MethodArgs) server.MethodResponse {
        return fmt.Sprintf("Hello %s", args[0].(string))
    })

    fmt.Println("RPC server is waiting for incoming requests...")
    userService.ServeForever()
}
