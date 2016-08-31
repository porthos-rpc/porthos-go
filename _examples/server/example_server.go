package main

import (
	"os"

	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/server"
	"github.com/porthos-rpc/porthos-go/status"
)

func main() {
	broker, err := server.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		log.Error("Error creating broker")
		panic(err)
	}

	defer broker.Close()

	// create the RPC server.
	userService, err := server.NewServer(broker, "UserService", server.Options{MaxWorkers: 40, AutoAck: false})

	// create and add the built-in metrics shipper.
	userService.AddExtension(server.NewMetricsShipperExtension(broker, server.MetricsShipperConfig{
		BufferSize: 100,
	}))

	if err != nil {
		log.Error("Error creating server")
		panic(err)
	}

	defer userService.Close()

	userService.Register("doSomething", func(req server.Request, res *server.Response) {
		// nothing to do yet.
	})

	userService.Register("doSomethingThatReturnsValue", func(req server.Request, res *server.Response) {
		type test struct {
			Original float64 `json:"original"`
			Sum      float64 `json:"sum"`
		}

		x := req.GetArg(0).AsFloat64()

		res.JSON(status.OK, test{x, x + 1})
	})

	log.Info("RPC server is waiting for incoming requests...")
	userService.ServeForever()
}
