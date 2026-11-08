module github.com/aidar-khabibullin/MSA_Rocket_Factory/bff

go 1.24.4

replace github.com/aidar-khabibullin/MSA_Rocket_Factory/shared => ../shared

require (
	github.com/aidar-khabibullin/MSA_Rocket_Factory/shared v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.73.0
)

require (
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
