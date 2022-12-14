package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/last/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SoftwareTransactionalMemoryServiceServer struct{}

type VariableItem struct {
	Variable string
}

type AssignmentItem struct {
	Variable             string   `protobuf:"bytes,1,opt,name=variable,proto3" json:"variable,omitempty"`
	Value                int32    `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

var db *mongo.Client
var assignmentdb *mongo.Collection
var mongoCtx context.Context
var sessions map[string]mongo.Session

func (s *SoftwareTransactionalMemoryServiceServer) CreateSession(ctx context.Context,
	req *services.CreateSessionRequest) (*services.CreateSessionResponse, error) {
	var session mongo.Session
	var err error

	if session, err = db.StartSession(); err != nil {
		fmt.Println(session)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start session"))
	}
	if err := session.StartTransaction(); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start a transaction"))
	}

	key := session.ID().String()
	sessions[key] = session

	response := &services.CreateSessionResponse{
		SessionID: &services.SessionID{
			Key: key,
		},
	}
	return response, nil
}

func (s *SoftwareTransactionalMemoryServiceServer) CommitSession(ctx context.Context,
	req *services.CommitSessionRequest) (*services.CommitSessionResponse, error) {

	sessionID := req.GetSessionID().GetKey()

	session, ok := sessions[sessionID]

	if !ok {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Session does not exist"))
	}

	var err error
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err = session.CommitTransaction(sc); err != nil {
			fmt.Println(err)
			return status.Errorf(codes.NotFound, fmt.Sprintf("Transaction could not be committed"))
		}
		return nil
	}); err != nil {
		fmt.Println(err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed during committing"))
	}
	return &services.CommitSessionResponse{}, nil
}

func (s *SoftwareTransactionalMemoryServiceServer) GetVariable(ctx context.Context,
	req *services.GetVariableRequest) (*services.GetVariableResponse, error) {

	sessionID := req.GetSessionID().GetKey()

	session, ok := sessions[sessionID]

	fmt.Println(ok)
	var err error

	if !ok {
		if session, err = db.StartSession(); err != nil {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start session"))
		}
		if err = session.StartTransaction(); err != nil {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start a transaction"))
		}
	}

	variable := req.GetVariable()
	data := AssignmentItem{}

	fmt.Printf("Received Get Request for variable: %s\n", variable)

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		result := assignmentdb.FindOne(sc, bson.M{"variable": variable})

		fmt.Print("got result in get:")
		fmt.Println(result)
		if err := result.Decode(&data); err != nil {
			sc.AbortTransaction(sc)
			fmt.Println("aborted")
			return status.Errorf(codes.NotFound, fmt.Sprintf("Could not find assignment with variable %s: %v", variable, err))
		}

		fmt.Println(err)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to execute operation in session while Getting"))
	}

	// result := assignmentdb.FindOne(ctx, bson.M{"variable": variable})
	// // Create an empty AssignmentItem to write our decode result to
	// data := AssignmentItem{}
	// // decode and write to data
	// if err := result.Decode(&data); err != nil {
	// 	return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find assignment with variable %s: %v", variable, err))
	// }

	response := &services.GetVariableResponse{
		Assignment: &services.Assignment{
			Variable: data.Variable,
			Value:    data.Value,
		},
	}

	return response, nil
}

func (s *SoftwareTransactionalMemoryServiceServer) SetVariable(ctx context.Context,
	req *services.SetVariableRequest) (*services.SetVariableResponse, error) {
	// Essentially doing req.Blog to access the struct with a nil check
	sessionID := req.GetSessionID().GetKey()
	assignment := req.GetAssignment()

	// fmt.Println(sessions)

	session, ok := sessions[sessionID]

	fmt.Println(ok)
	var err error

	if !ok {
		if session, err = db.StartSession(); err != nil {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start session"))
		}
		if err = session.StartTransaction(); err != nil {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to start a transaction"))
		}
	}

	fmt.Printf("Received Set Request for variable: %s to set value: %d\n", assignment.Variable, assignment.Value)

	// option to set variable if not found
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.M{"variable": assignment.Variable}
	update := bson.M{"$set": bson.M{"value": assignment.Value}}

	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// update the desired variable
		result := assignmentdb.FindOneAndUpdate(
			sc,
			filter,
			update,
			opts,
		)

		fmt.Println("Checking for errors in set")
		fmt.Println(err)
		// check for potential errors
		if result == nil {

			fmt.Println("FOUND AN ERROR, ABORTING")
			sc.AbortTransaction(sc)
			// return internal gRPC error to be handled later
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal error: %v", err),
			)
		}
		fmt.Println("No Errors")
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Failed to execute operation in session while Setting"))
	}

	if err != nil {
		return nil, err
	}

	resp := &services.SetVariableResponse{}

	return resp, nil
}

// func (s *SoftwareTransactionalMemoryServiceServer) GetVariable(ctx context.Context,
// 	req *services.GetVariableRequest) (*services.GetVariableResponse, error) {

// 	// Define the callback that specifies the sequence of operations to perform inside the transaction.
// 	// Important: You must pass sessCtx as the Context parameter to the operations for them to be executed in the transaction.
// 	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		data := AssignmentItem{}
// 		result := assignmentdb.FindOne(ctx, bson.M{"variable": req.GetVariable()})
// 		if err := result.Decode(&data); err != nil {
// 			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find assignment with Object Id %s: %v", req.GetVariable(), err))
// 		}
// 		response := &services.GetVariableResponse{
// 			Assignment: &services.Assignment{
// 				Variable: data.Variable,
// 				Value:    data.Value,
// 			},
// 		}
// 		return response, nil
// 	}

// 	// Start a session and run the callback using WithTransaction.
// 	session, err := db.StartSession()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer session.EndSession(ctx)
// 	response, err := session.WithTransaction(ctx, callback)

// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, ok := response.(*services.GetVariableResponse)

// 	if !ok {
// 		return nil, nil
// 	}

// 	return resp, nil

// }

// func (s *SoftwareTransactionalMemoryServiceServer) SetVariable(ctx context.Context,
// 	req *services.SetVariableRequest) (*services.SetVariableResponse, error) {
// 	// Essentially doing req.Blog to access the struct with a nil check
// 	assignment := req.GetAssignment()

// 	fmt.Printf("Received Set Request for variable: %s to set value: %d\n", assignment.Variable, assignment.Value)

// 	// option to set variable if not found
// 	opts := options.FindOneAndUpdate().SetUpsert(true)
// 	filter := bson.M{"variable": assignment.Variable}
// 	update := bson.M{"$set": bson.M{"value": assignment.Value}}

// 	// Define the callback that specifies the sequence of operations to perform inside the transaction.
// 	// Important: Must pass sessCtx as the Context parameter to the operations for them to be executed in the transaction.
// 	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

// 		// update the desired variable
// 		err := assignmentdb.FindOneAndUpdate(
// 			ctx,
// 			filter,
// 			update,
// 			opts,
// 		)

// 		// check for potential errors
// 		if err != nil {
// 			// return internal gRPC error to be handled later
// 			return nil, status.Errorf(
// 				codes.Internal,
// 				fmt.Sprintf("Internal error: %v", err),
// 			)
// 		}
// 		return &services.SetVariableResponse{}, nil
// 	}

// 	// Start a session and run the callback using WithTransaction.
// 	session, err := db.StartSession()

// 	if err != nil {
// 		return nil, err
// 	}
// 	defer session.EndSession(ctx)
// 	response, err := session.WithTransaction(ctx, callback)

// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, ok := response.(*services.SetVariableResponse)

// 	if !ok {
// 		return nil, nil
// 	}

// 	return resp, nil
// }

func main() {
	// Configure 'log' package to give file name and line number on eg. log.Fatal
	// Pipe flags to one another (log.LstdFLags = log.Ldate | log.Ltime)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Starting server on port :50051...")

	// Start our listener, 50051 is the default gRPC port
	listener, err := net.Listen("tcp", ":50051")
	// Handle errors if any
	if err != nil {
		log.Fatalf("Unable to listen on port :50051: %v", err)
	}

	// create sessions map
	sessions = make(map[string]mongo.Session)

	// Set options, here we can configure things like TLS support
	opts := []grpc.ServerOption{}
	// Create new gRPC server with (blank) options
	s := grpc.NewServer(opts...)
	// Create BlogService type
	srv := &SoftwareTransactionalMemoryServiceServer{}
	// Register the service with the server
	services.RegisterSoftwareTransactionalMemoryServiceServer(s, srv)

	// Initialize MongoDb client
	fmt.Println("Connecting to MongoDB...")

	// non-nil empty context
	mongoCtx = context.Background()

	// Connect takes in a context and options, the connection URI is the only option we pass for now
	db, err = mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://localhost:27017,localhost:27018/?replicaSet=rep"))
	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}

	// Check whether the connection was succesful by pinging the MongoDB server
	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	// Bind our collection to our global variable for use in other methods
	assignmentdb = db.Database("mydb").Collection("assignments")

	// Start the server in a child routine
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port :50051")

	// Right way to stop the server using a SHUTDOWN HOOK
	// Create a channel to receive OS signals
	c := make(chan os.Signal)

	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)

	// Block main routine until a signal is received
	// As long as user doesn't press CTRL+C a message is not passed and our main routine keeps running
	<-c

	// After receiving CTRL+C Properly stop the server
	fmt.Println("\nStopping the server...")
	s.Stop()
	listener.Close()
	fmt.Println("Closing MongoDB connection")
	db.Disconnect(mongoCtx)
	fmt.Println("Done.")
}
