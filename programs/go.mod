module github.com/last/programs

go 1.19

replace github.com/last/transactionalVariable => ../transactionalVariable

replace github.com/last/client => ../client

replace github.com/last/services => ../protos

require github.com/last/transactionalVariable v0.0.0-00010101000000-000000000000

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/last/client v0.0.0-00010101000000-000000000000 // indirect
	github.com/last/services v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
