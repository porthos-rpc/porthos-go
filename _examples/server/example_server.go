package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/log"
	"github.com/porthos-rpc/porthos-go/server"
	"github.com/porthos-rpc/porthos-go/status"
)

func main() {
	b, err := broker.NewBroker(os.Getenv("AMQP_URL"))
	defer b.Close()

	if err != nil {
		log.Error("Error creating broker")
		panic(err)
	}

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

	userService.Register("doSomething", func(req server.Request, res server.Response) {
		// nothing to do yet.
	})

	userService.Register("doSomethingElse", func(req server.Request, res server.Response) {
		m := make(map[string]int)
		_ = req.Bind(&m)
		log.Info("doSomethingElse with value %f", m["value"])
	})

	userService.Register("doSomethingThatReturnsValue", func(req server.Request, res server.Response) {
		type input struct {
			Value int `json:"value"`
		}

		type output struct {
			Original int `json:"original_value"`
			Sum      int `json:"value_plus_one"`
		}

		var i input

		_ = req.Bind(&i)

		res.JSON(status.OK, output{i.Value, i.Value + 1})
	})

	userService.ListenAndServe()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)

	<-sigc
	userService.Shutdown()
}
