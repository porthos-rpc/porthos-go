package main

import (
	"fmt"
	"os"
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

	for i := 0; i < 10000; i++ {
		ret, err := userService.Call("doSomethingThatReturnsValue").WithMap(map[string]interface{}{"value": i}).Sync()

		if err != nil {
			fmt.Println(err)
		} else {
			jsonResponse, _ := ret.UnmarshalJSON()

			fmt.Printf("Response %d. Original: %f. Sum: %f\n", i, jsonResponse["original_value"], jsonResponse["value_plus_one"])
		}
	}
}
