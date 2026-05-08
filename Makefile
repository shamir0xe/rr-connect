BINARY := rr-connect
MAIN   := main.go

.PHONY: build build-linux vet test

build:
	go build -o $(BINARY) $(MAIN)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 $(MAIN)

vet:
	go vet ./...

test:
	go test ./...
