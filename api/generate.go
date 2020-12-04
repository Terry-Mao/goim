package api

//go:generate protoc -I. -I$GOPATH/src --go-v1_out=plugins=grpc:. --go-v1_opt=paths=source_relative protocol/protocol.proto
//go:generate protoc -I. -I$GOPATH/src --go-v1_out=plugins=grpc:. --go-v1_opt=paths=source_relative comet/comet.proto
//go:generate protoc -I. -I$GOPATH/src --go-v1_out=plugins=grpc:. --go-v1_opt=paths=source_relative logic/logic.proto
