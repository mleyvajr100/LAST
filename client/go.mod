module last/client/main

go 1.19

require (
	google.golang.org/grpc v1.50.1
	last/services v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221114212237-e4508ebdbee1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace last/services => ../protos
