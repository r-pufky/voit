MODULE=$(shell go list -m)
VERSION=$(shell git describe --tags --always --dirty)
BINARY_NAME=voit
LDFLAGS=-ldflags "-X '$(MODULE)/cmd.Version=$(VERSION)'"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME)
