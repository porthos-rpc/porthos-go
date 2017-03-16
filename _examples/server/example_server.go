package main

import (
	"os"

	"github.com/porthos-rpc/porthos-go"
	"github.com/porthos-rpc/porthos-go/log"
)

func main() {
	b, err := porthos.NewBroker(os.Getenv("AMQP_URL"))
	defer b.Close()

	if err != nil {
		log.Error("Error creating broker")
		panic(err)
	}

	// create the RPC server.
	userService, err := porthos.NewServer(b, "UserService", porthos.Options{MaxWorkers: 40, AutoAck: false})

	if err != nil {
		log.Error("Error creating server")
		panic(err)
	}

	defer userService.Close()

	// create and add the built-in metrics shipper.
	userService.AddExtension(porthos.NewMetricsShipperExtension(b, porthos.MetricsShipperConfig{
		BufferSize: 100,
	}))

	// create and add the access log extension.
	userService.AddExtension(porthos.NewAccessLogExtension())

	userService.Register("doSomething", func(req porthos.Request, res porthos.Response) {
		// nothing to do yet.
	})

	userService.Register("doSomethingElse", func(req porthos.Request, res porthos.Response) {
		m := make(map[string]int)
		_ = req.Bind(&m)
		log.Info("doSomethingElse with value %f", m["value"])
	})

	userService.Register("doSomethingThatReturnsValue", func(req porthos.Request, res porthos.Response) {
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

	userService.ListenAndServe()
}
