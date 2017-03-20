// Package porthos is a RPC library for the Go programming language that operates over AMQP.
//
// Client
//
// The client is very simple. NewClient takes a broker, a service name and a timeout value (message TTL). The service name is only intended to serve as the request routing key (meaning every service name (or microservice) has its own queue). Each client declares only one response queue, in order to prevent broker's resources wastage.
//
// 		// first of all you need a broker
//	 	b, _ := porthos.NewBroker(os.Getenv("AMQP_URL"))
//	 	defer b.Close()
//
//	 	// then you create a new client (you can have as many clients as you want using the same broker)
//	 	calculatorService, _ := porthos.NewClient(b, "CalculatorService", 120)
//	 	defer calculatorService.Close()
//
//	 	// finally the remote call. It returns a response that contains the output channel.
//	 	ret, _ := calculatorService.Call("addOne").WithMap(map[string]interface{}{"value": 20}).Async()
//	 	defer ret.Dispose()
//
//	 	select {
//	 	case response := <-ret.ResponseChannel():
//	 	    jsonResponse, _ := response.UnmarshalJSON()
//
// 	    fmt.Printf("Original: %f, sum: %f\n", jsonResponse["original_value"], jsonResponse["value_plus_one"])
// 		case <-time.After(2 * time.Second):
// 	    	fmt.Println("Timed out :(")
// 		}
//
// Server
//
// The server also takes a broker and a service name. After that, you Register all your handlers and finally ListenAndServe.
//
//		b, _ := porthos.NewBroker(os.Getenv("AMQP_URL"))
//		defer b.Close()
//
//		calculatorService, _ := porthos.NewServer(b, "CalculatorService", 10, false)
//		defer calculatorService.Close()
//
//		calculatorService.Register("addOne", func(req porthos.Request, res *porthos.Response) {
//		    type input struct {
//		        Value int `json:"value"`
//		    }
//
//		    type output struct {
//		        Original int `json:"original_value"`
//		        Sum      int `json:"value_plus_one"`
//		    }
//
//		    var i input
//
//		    _ = req.Bind(&i)
//
//		    res.JSON(porthos.OK, output{i.Value, i.Value + 1})
//		})
//
//		calculatorService.Register("subtract", func(req porthos.Request, res *porthos.Response) {
//		    // subtraction logic here...
//		})
//
//		calculatorService.ListenAndServe()
//
// Extensions
//
// Extensions can be used to add custom actions to the RPC Server. The available "events" are incoming and outgoing.
//
//     func NewLoggingExtension() *Extension {
//         ext := porthos.NewExtension()
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
//     userService.AddExtension(porthos.NewMetricsShipperExtension(broker, porthos.MetricsShipperConfig{
//         BufferSize: 150,
//     }))
//
// Access Log Extension
//
//     userService.AddExtension(NewAccessLogExtension())
//
package porthos
