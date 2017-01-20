// Package porthos is a RPC library for the Go programming language that operates over AMQP.
//
// Client
//
// The client is very simple. NewClient takes a broker, a service name and a timeout value (message TTL). The service name is only intended to serve as the request routing key (meaning every service name (or microservice) has its own queue). Each client declares only one response queue, in order to prevent broker's resources wastage.
//
//     // first of all you need a broker
//     b, _ := broker.NewBroker(os.Getenv("AMQP_URL"))
//     defer b.Close()
//
//     // then you create a new client (you can have as many clients as you want using the same broker)
//     calculatorService, _ := client.NewClient(b, "CalculatorService", 120)
//     defer calculatorService.Close()
//
//     // finally the remote call. It returns a response that contains the output channel.
//     ret, _ := calculatorService.Call("addOne", 10)
//     defer ret.Dispose()
//
//     select {
//     case response := <-ret.ResponseChannel():
//         jsonResponse, _ := response.UnmarshalJSON()
//
//         fmt.Printf("Original: %f, sum: %f\n", jsonResponse["original"], jsonResponse["sum"])
//     case <-time.After(2 * time.Second):
//         fmt.Println("Timed out :(")
//     }
//
// Server
//
// The server also takes a broker and a service name. After that, you Register all your handlers and finally ServeForever.
//
//     b, _ := broker.NewBroker(os.Getenv("AMQP_URL"))
//     defer b.Close()
//
//     calculatorService, _ := server.NewServer(b, "CalculatorService", 10, false)
//     defer calculatorService.Close()
//
//     calculatorService.Register("addOne", func(req server.Request, res *server.Response) {
//         type response struct {
//             Original    float64 `json:"original"`
//             Sum         float64 `json:"sum"`
//         }
//
//         x := req.GetArg(0).AsFloat64()
//
//         res.JSON(status.OK, response{x, x+1})
//     })
//
//     calculatorService.Register("subtract", func(req server.Request, res *server.Response) {
//         // subtraction logic here...
//     })
//
//     fmt.Println("RPC server is waiting for incoming requests...")
//     b.WaitUntilConnectionCloses()
//
// Extensions
//
// Extensions can be used to add custom actions to the RPC Server. The available "events" are incoming and outgoing.
//
//     import "github.com/porthos-rpc/porthos-go/server"
//
//     func NewLoggingExtension() *Extension {
//         ext := server.NewExtension()
//
//         go func() {
//             for {
//                 select {
//                 case in := <-ext.Incoming():
//                     log.Info("Before executing method: %s", in.Request.MethodName)
//                 case out := <-ext.Outgoing():
//                     log.Info("After executing method: %s", out.Request.MethodName)
//                 }
//             }
//         }()
//
//         return ext
//     }
//
// Then you just have to add the extension to the server:
//
//     userService.AddExtension(NewLoggingExtension())
//
// Built-in extensions
//
// Metrics Shipper Extension
//
// This extension will ship metrics to the AMQP broker, any application can consume and display them as needed.
//
//     userService.AddExtension(rpc.NewMetricsShipperExtension(broker, rpc.MetricsShipperConfig{
//         BufferSize: 150,
//     }))
//
// Access Log Extension
//
//     userService.AddExtension(NewAccessLogExtension())
//
package porthos
