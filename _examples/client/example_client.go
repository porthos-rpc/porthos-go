package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/porthos-rpc/porthos-go/broker"
	"github.com/porthos-rpc/porthos-go/client"
)

func main() {
	b, err := broker.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		fmt.Printf("Error creating broker")
		panic(err)
	}

	defer b.Close()

	// create a client with a default timeout of 1 second.
	userService, err := client.NewClient(b, "UserService", 1000)

	if err != nil {
		fmt.Printf("Error creating client")
		panic(err)
	}

	defer userService.Close()

	// call a remote method that is "void".
	userService.CallVoid("doSomething", 20)
	fmt.Println("Service userService.doSomething invoked")

	var wg sync.WaitGroup

	// call a lot of methods concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(idx int) {
			ret, _ := userService.Call("doSomethingThatReturnsValue", idx)
			defer ret.Dispose()

			select {
			case response := <-ret.ResponseChannel():
				jsonResponse, _ := response.UnmarshalJSON()

				fmt.Printf("Response %d. Original: %f. Sum: %f\n", idx, jsonResponse["original"], jsonResponse["sum"])
			case <-time.After(2 * time.Second):
				fmt.Printf("Timed out %d :(\n", idx)
			}
			wg.Done()
		}(i)
	}

	// wait (to give time to execute all goroutines)
	wg.Wait()
}
