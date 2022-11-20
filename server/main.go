package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"last/services"

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

func (s *SoftwareTransactionalMemoryServiceServer) GetVariable(ctx context.Context,
	req *services.GetVariableRequest) (*services.GetVariableResponse, error) {
	// convert string id (from proto) to mongoDB ObjectId
	// oid, err := primitive.ObjectIDFromHex(req.GetVariable())
	// if err != nil {
	// 	return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	// }

	result := assignmentdb.FindOne(ctx, bson.M{"variable": req.GetVariable()})
	// Create an empty AssignmentItem to write our decode result to
	data := AssignmentItem{}
	// decode and write to data
	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find assignment with Object Id %s: %v", req.GetVariable(), err))
	}
	// Cast to ReadBlogRes type
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
	assignment := req.GetAssignment()

	fmt.Printf("Received Set Request for variable: %s to set svalue: %d\n", assignment.Variable, assignment.Value)
	// Now we have to convert this into a BlogItem type to convert into BSON
	// data := AssignmentItem{
	// 	Variable: assignment.Variable,
	// 	Value:    assignment.Value,
	// }

	// option to set variable if not found

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.M{"variable": assignment.Variable}
	// filter := bson.D{"v": 8}
	update := bson.M{"$set": bson.M{"value": assignment.Value}}
	// var updatedDocument bson.M
	err := assignmentdb.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	)

	// Insert the data into the database, result contains the newly generated Object ID for the new document
	// _, err := assignmentdb.InsertOne(mongoCtx, data)
	// check for potential errors
	if err != nil {
		// return internal gRPC error to be handled later
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	// return the blog in a CreateBlogRes type
	return &services.SetVariableResponse{}, nil
}

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
	db, err = mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
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
