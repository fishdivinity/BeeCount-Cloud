module github.com/fishdivinity/BeeCount-Cloud/services/log

go 1.25.6

require (
	github.com/fishdivinity/BeeCount-Cloud/common v0.0.0
	github.com/rs/zerolog v1.34.0
	google.golang.org/grpc v1.78.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260122232226-8e98ce8d340d // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/fishdivinity/BeeCount-Cloud/common => ../../common
