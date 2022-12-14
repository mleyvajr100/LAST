package client

import (
	"context"
	"log"
	"time"

	"github.com/last/services"

	"google.golang.org/grpc"
)

// Client and context global vars
var txClient services.SoftwareTransactionalMemoryServiceClient
var requestCtx context.Context
var requestOpts grpc.DialOption

func clientConnectionSetup() {
	requestCtx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	// Establish insecure grpc options (no TLS)
	requestOpts = grpc.WithInsecure()
	// Dial the server, returns a client connection
	conn, err := grpc.Dial("localhost:50051", requestOpts)

	if err != nil {
		log.Fatalf("Unable to establish client connection to localhost:50051: %v", err)
	}
	// Instantiate the TransactionalMemoryServiceClient with our client connection to the server
	txClient = services.NewSoftwareTransactionalMemoryServiceClient(conn)
}

func CreateSession() string {
	createReq := &services.CreateSessionRequest{}

	if txClient == nil {
		clientConnectionSetup()
	}
	resp, err := txClient.CreateSession(requestCtx, createReq)

	if err != nil {
		log.Fatal(err)
	}

	return resp.GetSessionID().Key
}

func CommitSession(sessionID string) {
	commitReq := &services.CommitSessionRequest{
		SessionID: &services.SessionID{
			Key: sessionID,
		},
	}

	if txClient == nil {
		clientConnectionSetup()
	}
	_, err := txClient.CommitSession(requestCtx, commitReq)

	if err != nil {
		log.Fatal(err)
	}
}

func SetVariable(variable string, value int32, sessionID string) {
	setReq := &services.SetVariableRequest{
		Assignment: &services.Assignment{
			Variable: variable,
			Value:    value,
		},
		SessionID: &services.SessionID{
			Key: sessionID,
		},
	}

	if txClient == nil {
		clientConnectionSetup()
	}

	txClient.SetVariable(requestCtx, setReq)
}

func GetVariable(variable string, sessionID string) int32 {
	getReq := &services.GetVariableRequest{
		Variable: variable,
		SessionID: &services.SessionID{
			Key: sessionID,
		},
	}

	if txClient == nil {
		clientConnectionSetup()
	}

	resp, err := txClient.GetVariable(requestCtx, getReq)

	if err != nil {
		log.Fatal(err)
	}

	return resp.GetAssignment().Value
}
