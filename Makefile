BINARY_NAME=voit
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X 'main.Version=$(VERSION)'"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME)
