module github.com/fishdivinity/BeeCount-Cloud/services/beecount

go 1.25.6

require (
	github.com/fishdivinity/BeeCount-Cloud/common v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.10.2
	google.golang.org/grpc v1.78.0
)

replace github.com/fishdivinity/BeeCount-Cloud/common => ../../common

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260122232226-8e98ce8d340d // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
