package main

import (
	"os"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/server"
	"github.com/porthos-rpc/porthos-go/status"
)

func main() {
	b, err := broker.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		log.Error("Error creating broker")
		panic(err)
	}

	defer b.Close()

	// create the RPC server.
	userService, err := server.NewServer(b, "UserService", server.Options{MaxWorkers: 40, AutoAck: false})

	// create and add the built-in metrics shipper.
	userService.AddExtension(server.NewMetricsShipperExtension(b, server.MetricsShipperConfig{
		BufferSize: 100,
	}))

	// create and add the access log extension.
	userService.AddExtension(server.NewAccessLogExtension())

	if err != nil {
		log.Error("Error creating server")
		panic(err)
	}

	defer userService.Close()

	userService.Register("doSomething", func(req server.Request, res server.Response) {
		// nothing to do yet.
	})

	userService.Register("doSomethingElse", func(req server.Request, res server.Response) {
		form, _ := req.MapForm()
		x, _ := form.GetArg("someField").AsFloat64()

		log.Info("doSomethingElse with someField %f", x)

	})

	userService.Register("doSomethingThatReturnsValue", func(req server.Request, res server.Response) {
		type test struct {
			Original float64 `json:"original"`
			Sum      float64 `json:"sum"`
		}

		form, _ := req.IndexForm()
		x, _ := form.GetArg(0).AsFloat64()

		res.JSON(status.OK, test{x, x + 1})
	})

	userService.Start()

	log.Info("RPC server is waiting for incoming requests...")
	b.WaitUntilConnectionCloses()
}
