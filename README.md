# Porthos [![GoDoc](https://godoc.org/github.com/porthos-rpc/porthos-go?status.svg)](http://godoc.org/github.com/porthos-rpc/porthos-go) [![Build Status](https://travis-ci.org/porthos-rpc/porthos-go.svg?branch=master)](https://travis-ci.org/porthos-rpc/porthos-go) [![Go Report Card](https://goreportcard.com/badge/github.com/porthos-rpc/porthos-go)](https://goreportcard.com/report/github.com/porthos-rpc/porthos-go) [![License](https://img.shields.io/github/license/porthos-rpc/porthos-go.svg?maxAge=2592000)]()

A RPC library for the Go programming language that operates over AMQP.

## Status

Beta. Server and Client API may change a bit.

## Goal

Provide a language-agnostic RPC library to write distributed systems.

## Client

The client is very simple. `NewClient` takes a broker, a `service name` and a timeout value (message TTL). The `service name` is only intended to serve as the request `routing key` (meaning every `service name` (or microservice) has its own queue). Each client declares only one `response queue`, in order to prevent broker's resources wastage.

### Creating a new client

```go
// first of all you need a broker
b, _ := porthos.NewBroker(os.Getenv("AMQP_URL"))
defer b.Close()

// then you create a new client (you can have as many clients as you want using the same broker)
calculatorService, _ := porthos.NewClient(b, "CalculatorService", 120)
defer calculatorService.Close()
```

### The Call builder

#### `.Call(methodName string)`
Creates a call builder.

#### `.WithTimeout(d duration)`
Defines a timeout for the current call. Example:

```go
calculatorService.Call("addOne").WithTimeout(2*time.Second)...
```

#### `.WithMap(m map[string]interface{})`
Sets the given map as the request body of the current call. The content type used is `application/json`. Example:

```go
calculatorService.Call("addOne").WithMap(map[string]interface{}{"value": 20})...
```

#### `.WithStruct(s interface{})`
Sets the given struct as the request body of the current call. The content type used is `application/json`. Example:

```go
calculatorService.Call("addOne").WithStruct(myStruct)...
```

#### `.WithArgs(args ...interface{})`
Sets the given args as the request body of the current call. The content type used is `application/json`. Example:

```go
calculatorService.Call("add").WithArgs(1, 2)...
```

#### `.WithBody(body []byte)`
Sets the given byte array as the request body of the current call. The content type is `application/octet-stream`. Example:

```go
calculatorService.Call("addOne").WithBody(byteArray)...
```

#### `.WithBodyContentType(body []byte, contentType string)`
Sets the given byte array as the request body of the current call. Also takes a contentType. Example:

```go
calculatorService.Call("addOne").WithBodyContentType(jsonByteArrayJ, "application/json")...
```

#### `.Async() (Slot, error)`
Performs the remote call and returns a slot that contains the response `channel`. Example:

```go
s, err := calculatorService.Call("addOne").WithArgs(1).Async()
s.Dispose()

r := <-s.ResponseChannel()
json, err := r.UnmarshalJSON()
```

You can easily handle timeout with a `select`:

```go
select {
case r := <-s.ResponseChannel():
    json, err := r.UnmarshalJSON()
case <-time.After(2 * time.Second):
    ...
}
```

#### `.Sync() (*ClientResponse, error)`
Performs the remote call and returns the response. Example:

```go
r, err := calculatorService.Call("addOne").WithMap(map[string]interface{}{"value": 20}).Sync()
json, err := r.UnmarshalJSON()
```

#### `.Void() error`
Performs the remote call that doesn't return anything. Example:

```go
err := loggingService.Call("log").WithArgs("INFO", "some log message").Void()
```

You can find a full client example at `_examples/client/example_client.go`.

## Server

The server also takes a broker and a `service name`. After that, you `Register` all your handlers and finally `ServeForever`.

### Creating a new server

```go
b, _ := porthos.NewBroker(os.Getenv("AMQP_URL"))
defer b.Close()

calculatorService, _ := porthos.NewServer(b, "CalculatorService", 10, false)
defer calculatorService.Close()
```

#### `.Register(methodName string, handler MethodHandler)`
Register a method with the given handler. Example:

```go
calculatorService.Register("addOne", func(req porthos.Request, res *porthos.Response) {
    type input struct {
        Value int `json:"value"`
    }

    type output struct {
        Original int `json:"original_value"`
        Sum      int `json:"value_plus_one"`
    }

    var i input

    _ = req.Bind(&i)

    res.JSON(porthos.OK, output{i.Value, i.Value + 1})
})
```

#### `.RegisterWithSpec(method string, handler MethodHandler, spec Spec)`
Register a method with the given handler and a `Spec`. Example:

```go
calculatorService.RegisterWithSpec("addOne", addOneHandler, porthos.Spec{
    Description: "Adds one to the given int argument",
    Request: porthos.ContentSpec{
        ContentType: "application/json",
        Body:        porthos.BodySpecFromStruct(input{}),
    },
    Response: porthos.ContentSpec{
        ContentType: "application/json",
        Body:        porthos.BodySpecFromArray(output{}),
    },
})
```

Through the [Specs Shipper Extension](#specs-shipper-extension) the specs are shipped to a queue call `porthos.specs` and can be displayed in the [Porthos Playground](https://github.com/porthos-rpc/porthos-playground).

#### `.AddExtension(ext Extension)`
Adds the given extension to the server.

#### `.ListenAndServe()`
Starts serving RPC requests.

```go
calculatorService.ListenAndServe()
```

#### `.Close()`
Close the server and AMQP channel. This method returns right after the AMQP channel is closed. In order to give time to the current request to finish (if there's one) it's up to you to wait using the NotifyClose.

#### `.Shutdown()`
Shutdown shuts down the server and AMQP channel. It provider graceful shutdown, since it will wait the result of <-s.NotifyClose().

You can find a full server example at `_examples/server/example_server.go`.

## Extensions

Extensions can be used to add custom actions to the RPC Server. The available "events" are `incoming` and `outgoing`.

```go
func NewLoggingExtension() *Extension {
    ext := porthos.NewExtension()

    go func() {
        for {
            select {
            case in := <-ext.Incoming():
                log.Printf("Before executing method: %s", in.Request.MethodName)
            case out := <-ext.Outgoing():
                log.Printf("After executing method: %s", out.Request.MethodName)
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

### Built-in extensions

#### Metrics Shipper Extension

This extension will ship metrics to the AMQP broker, any application can consume and display them as needed.

```go
userService.AddExtension(porthos.NewMetricsShipperExtension(broker, porthos.MetricsShipperConfig{
    BufferSize: 150,
}))
```

#### Access Log Extension

```go
userService.AddExtension(NewAccessLogExtension())
```

#### Specs Shipper Extension

```go
userService.AddExtension(porthos.NewSpecShipperExtension(broker))
```

## Contributing
Please read the [contributing guide](CONTRIBUTING.md)

Pull requests are very much welcomed. Make sure a test or example is included that covers your change.

Docker is being used for the local environment. To build/run/test your code you can bash into the server container:

```sh
$ docker-compose run server bash
root@porthos:/go/src/github.com/porthos-rpc/porthos-go# go run example_client.go
```
