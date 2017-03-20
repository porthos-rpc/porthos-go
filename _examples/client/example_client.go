package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	porthos "github.com/porthos-rpc/porthos-go"
)

func main() {
	b, err := porthos.NewBroker(os.Getenv("AMQP_URL"))

	if err != nil {
		fmt.Printf("Error creating broker")
		panic(err)
	}

	defer b.Close()

	// create a client with a default timeout of 1 second.
	userService, err := porthos.NewClient(b, "UserService", 1*time.Second)

	if err != nil {
		fmt.Printf("Error creating client")
		panic(err)
	}

	defer userService.Close()

	// call a remote method that is "void".
	userService.Call("doSomething").WithMap(map[string]interface{}{"value": 20}).Void()
	fmt.Println("Service userService.doSomething invoked")

	// call a remote method that returns the value right away.
	response, err := userService.Call("doSomethingThatReturnsValue").WithMap(map[string]interface{}{"value": 20}).Sync()
	jsonResponse, _ := response.UnmarshalJSON()
	fmt.Println("Service userService.doSomethingThatReturnsValue sync call: %d", jsonResponse["value_plus_one"])

	var wg sync.WaitGroup

	// call a lot of methods concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(idx int) {
			ret, _ := userService.Call("doSomethingThatReturnsValue").WithMap(map[string]interface{}{"value": idx}).Async()
			defer ret.Dispose()

			select {
			case response := <-ret.ResponseChannel():
				jsonResponse, _ := response.UnmarshalJSON()

				fmt.Printf("Response %d. Original: %f. Sum: %f\n", idx, jsonResponse["original_value"], jsonResponse["value_plus_one"])
			case <-time.After(2 * time.Second):
				fmt.Printf("Timed out %d :(\n", idx)
			}
			wg.Done()
		}(i)
	}

	// wait (to give time to execute all goroutines)
	wg.Wait()

	// call a remote method that returns the value right away.
	response, _ = userService.Call("doSomethingElse").WithMap(porthos.Map{"someField": 10}).Sync()
	jsonResponse, _ = response.UnmarshalJSON()
	fmt.Println("Service userService.doSomethingThatReturnsValue sync call: %d", jsonResponse["value_plus_one"])

}
