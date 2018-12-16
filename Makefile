# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	cp cmd/comet/comet.toml target/comet.toml
	cp cmd/logic/logic.toml target/logic.toml
	cp cmd/job/job.toml target/job.toml
	$(GOBUILD) -o target/comet cmd/comet/main.go
	$(GOBUILD) -o target/logic cmd/logic/main.go
	$(GOBUILD) -o target/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf target

run-comet:
	target/comet -conf=target/comet.toml -region=sh -zone=sh001 deploy.env=dev weight=10 addrs=127.0.0.1

run-logic:
	target/logic -conf=target/logic.toml -region=sh -zone=sh001 deploy.env=dev weight=10

run-job:
	target/job -conf=target/job.toml -region=sh -zone=sh001 deploy.env=dev
