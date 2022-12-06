# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.


GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

privacy:
	go build pkg/types/*.go
	go build pkg/variables/*.go
	go build pkg/wires/*.go
	go build pkg/garbler/*.go
	go build pkg/circuit/*.go
	go build pkg/engine/*.go
	go build pkg/compiler/*.go
	go build pkg/builder/*.go
	go build pkg/interpreter/*.go
	go build pkg/variables/*.go
	go build pkg/runner/*.go
	go build *.go
	@echo "Building ixxo-privacy"
