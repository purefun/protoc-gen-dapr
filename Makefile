PATH := $(PATH):$(pwd)

.PHONY: install build protoc gen

install:
	brew install protobuf
	protoc --version
	go get google.golang.org/protobuf/cmd/protoc-gen-go
	protoc-gen-go --version

build:
	go build ./cmd/protoc-gen-dapr/.

protoc:
	protoc --go_out=. --dapr_out=. examples/echo.proto --experimental_allow_proto3_optional

gen: build protoc
