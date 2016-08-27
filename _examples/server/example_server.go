package main

import (
	"os"

	"github.com/porthos-rpc/porthos-go/log"
	rpc "github.com/porthos-rpc/porthos-go/server"
)

func main() {
	broker, err := rpc.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		log.Error("Error creating broker")
		panic(err)
	}

	defer broker.Close()

	// create the RPC server.
	userService, err := rpc.NewServer(broker, "UserService", rpc.ServerOptions{MaxWorkers: 40, AutoAck: false})

	// create and add the built-in metrics shipper.
	userService.AddExtension(rpc.NewMetricsShipperExtension(broker, rpc.MetricsShipperConfig{
		BufferSize: 100,
	}))

	if err != nil {
		log.Error("Error creating server")
		panic(err)
	}

	defer userService.Close()

	userService.Register("doSomething", func(req rpc.Request, res *rpc.Response) {
		// nothing to do yet.
	})

	userService.Register("doSomethingThatReturnsValue", func(req rpc.Request, res *rpc.Response) {
		type test struct {
			Original float64 `json:"original"`
			Sum      float64 `json:"sum"`
		}

		x := req.GetArg(0).AsFloat64()

		res.JSON(test{x, x + 1})
	})

	log.Info("RPC server is waiting for incoming requests...")
	userService.ServeForever()
}
