# Porthos

A RPC library for the Go programming language that operates over AMQP.

## Client

The client is very simple. `NewClient` takes a broker, a `service name` and a timeout value (message TTL). The `service name` is only intended to serve as the request `routing key` (meaning every `service name` (or microservice) has its own queue).

Each `Call` returns two channels: one for the actual response and other for signaling a timeout. The response depends on the ContentType returned by the remote procedure. In case of an `application/json` response, a map will be returned, otherwise you will get a string.

Each client instance has its own `response queue`. To match requests and responses there's a slot array. Each `Call` uses a free slot (`O(n)`) and when the response comes, we get the slot by index (`O(1)`), since the correlationId contains the index. So the request and response operation is `O(n)`+ `O(1)`.

```go
// first of all you need a broker
broker, _ := client.NewBroker(os.Getenv("AMQP_URL"))
defer broker.Close()

// then you create a new client (you can have as many clients as you want using the same broker)
userService, _ := client.NewClient(broker, "UserService", 120)
defer userService.Close()

// finally the remote call. It returns two channels.
response, timeout := userService.Call("doSomethingThatReturnsValue", 20)

// handle the actual response and timeout.
select {
case res := <- response:
    fmt.Printf("Response: %#v\n", res)
case <- timeout:
    fmt.Println("Timed out :(")
}
```

## Server

Work in progress...
