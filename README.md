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
response, _ := calculatorService.Call("addOne", 10)
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
import (
	"github.com/porthos-rpc/porthos-go/server"
	"github.com/porthos-rpc/porthos-go/status"
)

broker, _ := server.NewBroker(os.Getenv("AMQP_URL"))
defer broker.Close()

calculatorService, _ := server.NewServer(broker, "CalculatorService", 10, false)
defer calculatorService.Close()

calculatorService.Register("addOne", func(req server.Request, res *server.Response) {
    type response struct {
        Original    float64 `json:"original"`
        Sum         float64 `json:"sum"`
    }

    x := req.GetArg(0).AsFloat64()

    res.JSON(status.OK, response{x, x+1})
})

calculatorService.Register("subtract", func(req server.Request, res *server.Response) {
    // subtraction logic here...
})

fmt.Println("RPC server is waiting for incoming requests...")
calculatorService.ServeForever()
```

### Extensions

Extensions can be used to add custom actions to the RPC Server. The available "events" are `incoming` and `outgoing`.

```go
import "github.com/porthos-rpc/porthos-go/server"

func NewLoggingExtension() *Extension {
	ext := server.NewExtension()

	go func() {
		for {
			select {
			case in := <-ext.Incoming():
				log.Info("Before executing method: %s", in.Request.MethodName)
			case out := <-ext.Outgoing():
				log.Info("After executing method: %s", out.Request.MethodName)
			}
		}
	}()

	return ext
}
```

Then you just have to add the extension to the server:

```go
userService.AddExtension(NewLoggingExtension())
```

#### Built-in extensions

Currently there's only the "Metrics Shipper Extension". You can use it like:

```go
userService.AddExtension(rpc.NewMetricsShipperExtension(broker, rpc.MetricsShipperConfig{
	FlushInterval: 30 * time.Second,
}))
```

## Contributing

Pull requests are very much welcomed. Make sure a test or example is included that covers your change.

Docker is being used for the local environment. To build/run/test your code you can bash into the server container:

```sh
$ docker-compose run server bash
root@porthos:/go/src/github.com/porthos-rpc/porthos-go# go run example_client.go
```
