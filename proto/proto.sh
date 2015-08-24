#/bin/bash
protoc_path=/Users/maojian/Work/programfiles/protobuf/src/
gogoproto_path=/Users/maojian/Work/github/go/src/github.com/

# comet
protoc ./comet/comet.proto --gofast_out=./comet/ --proto_path=${protoc_path} --proto_path=${gogoproto_path} --proto_path=./comet/

# router
protoc ./router/router.proto --gofast_out=./router/ --proto_path=${protoc_path} --proto_path=${gogoproto_path} --proto_path=./router/

# logic
protoc ./logic/logic.proto --gofast_out=./logic/ --proto_path=${protoc_path} --proto_path=${gogoproto_path} --proto_path=./logic/
