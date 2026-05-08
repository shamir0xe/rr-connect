BINARY  := rr-connect
MAIN    := main.go
MODULE  := github.com/shamir0xe/rr-connect

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "\
  -s -w \
  -X $(MODULE)/src/build.Version=$(VERSION) \
  -X $(MODULE)/src/build.Commit=$(COMMIT) \
  -X $(MODULE)/src/build.Date=$(DATE)"

.PHONY: build build-linux vet test

build:
	go build $(LDFLAGS) -o $(BINARY) $(MAIN)

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 $(MAIN)

vet:
	go vet ./...

test:
	go test -v -count=1 ./...
