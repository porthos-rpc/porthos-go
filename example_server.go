package main

import (
	"fmt"
	"os"

	"github.com/gfronza/porthos/server"
)

func main() {
	broker, err := server.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		fmt.Printf("Error creating broker")
		panic(err)
	}

	defer broker.Close()

	userService, err := server.NewServer(broker, "UserService")

	if err != nil {
		fmt.Printf("Error creating server")
		panic(err)
	}

	defer userService.Close()

	// userService.Register("doSomethingThatReturnsValue", func(intValue int) int {
	// 	return intValue + 1
	// })
}
