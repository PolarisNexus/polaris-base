module github.com/PolarisNexus/polaris-base/services/platform-admin

go 1.23.4

require (
	connectrpc.com/connect v1.18.1
	github.com/PolarisNexus/polaris-base/api/gen/go v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.34.0
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.1 // indirect
)

replace github.com/PolarisNexus/polaris-base/api/gen/go => ../../api/gen/go
