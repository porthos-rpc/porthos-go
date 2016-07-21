package main

import (
    "fmt"
    "os"

    rpc "github.com/gfronza/porthos/server"
)

func main() {
    broker, err := rpc.NewBroker(os.Getenv("AMQP_URL"))

    if err != nil {
        fmt.Printf("Error creating broker")
        panic(err)
    }

    defer broker.Close()

    userService, err := rpc.NewServer(broker, "UserService", 4)

    if err != nil {
        fmt.Printf("Error creating server")
        panic(err)
    }

    defer userService.Close()

    userService.Register("doSomething", func(req rpc.Request, res *rpc.Response) {
        // nothing to do yet.
    })

    userService.Register("doSomethingThatReturnsValue", func(req rpc.Request, res *rpc.Response) {
        type test struct {
            Original    float64 `json:"original"`
            Sum         float64 `json:"sum"`
        }

        x := req.GetArg(0).AsFloat64()

        res.JSON(test{x, x+1})
    })

    fmt.Println("RPC server is waiting for incoming requests...")
    userService.ServeForever()
}
