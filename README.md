# Porthos

A RPC library for the Go programming language that operates over AMQP.

## Status

Beta. Server and Client API may change a bit.

## Goal

Provide a language-agnostic RPC library to write distributed systems.

## Client

The client is very simple. `NewClient` takes a broker, a `service name` and a timeout value (message TTL). The `service name` is only intended to serve as the request `routing key` (meaning every `service name` (or microservice) has its own queue). Each client declares only one `response queue`, in order to prevent broker's resources wastage.


```go
// first of all you need a broker
broker, _ := client.NewBroker(os.Getenv("AMQP_URL"))
defer broker.Close()

// then you create a new client (you can have as many clients as you want using the same broker)
calculatorService, _ := client.NewClient(broker, "CalculatorService", 120)
defer calculatorService.Close()

// finally the remote call. It returns a response that contains the output channel.
response := calculatorService.Call("addOne", 10)
defer response.Dispose()

select {
case res := <-response.Out():
    var jsonResponse map[string]interface{}
    json.Unmarshal(res, &jsonResponse)

    fmt.Printf("Original: %f, sum: %f\n", jsonResponse["original"], jsonResponse["sum"])
case <-time.After(2 * time.Second):
    fmt.Println("Timed out :(")
}
```

## Server

The server also takes a broker and a `service name`. After that, you `Register` all your handlers and finally `ServeForever`.

```go
broker, _ := rpc.NewBroker(os.Getenv("AMQP_URL"))
defer broker.Close()

calculatorService, _ := rpc.NewServer(broker, "CalculatorService")
defer calculatorService.Close()

calculatorService.Register("addOne", func(req rpc.Request, res *rpc.Response) {
    type response struct {
        Original    float64 `json:"original"`
        Sum         float64 `json:"sum"`
    }

    x := req.GetArg(0).AsFloat64()

    res.JSON(response{x, x+1})
})

calculatorService.Register("subtract", func(req rpc.Request, res *rpc.Response) {
    // subtraction logic here...
})

fmt.Println("RPC server is waiting for incoming requests...")
calculatorService.ServeForever()
```

## Acknowledgements

A special thanks to my coworker https://github.com/skrater.

## Contributing

Pull requests are very much welcomed. Make sure a test or example is included that covers your change.

Docker is being used for the local environment. To build/run/test your code you can bash into the server container:

```sh
$ docker-compose run server bash
root@porthos:/go/src/github.com/gfronza/porthos# go run example_client.go
```
