package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"last/services"

	"google.golang.org/grpc"
)

// Client and context global vars
var client services.SoftwareTransactionalMemoryServiceClient
var requestCtx context.Context
var requestOpts grpc.DialOption

func main() {
	fmt.Println("Starting Transactional Memory Service Client")

	// Establish context to timeout if server does not respond
	requestCtx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	// Establish insecure grpc options (no TLS)
	requestOpts = grpc.WithInsecure()
	// Dial the server, returns a client connection
	conn, err := grpc.Dial("localhost:50051", requestOpts)

	if err != nil {
		log.Fatalf("Unable to establish client connection to localhost:50051: %v", err)
	}

	// Instantiate the BlogServiceClient with our client connection to the server
	client = services.NewSoftwareTransactionalMemoryServiceClient(conn)

	variable := "z"
	var value int32 = 256

	fmt.Println("Creating SetVariable Request")
	setReq := &services.SetVariableRequest{
		Assignment: &services.Assignment{
			Variable: variable,
			Value:    value,
		},
	}

	fmt.Printf("Set variable: %s to value: %d\n", variable, value)

	// fmt.Println("Creating GetVariable Request")
	// getReq := &services.GetVariableRequest{
	// 	Variable: "x",
	// }

	// fmt.Println("Sending GetVariable Request")

	// start := time.Now()
	// for i := 0; i < 10; i++ {
	// 	client.GetVariable(requestCtx, getReq)
	// }

	// // Code to measure
	// duration := time.Since(start)

	// // Formatted string, such as "2h3m0.5s" or "4.503μs"
	// fmt.Println(duration)

	// resp.GetAssignment()
	// fmt.Println(resp)
	// value := resp.GetAssignment().Value
	// fmt.Printf("variable x has value: %d", value)

	client.SetVariable(requestCtx, setReq)
}
