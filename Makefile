BIN_NAME=kustomize

export GO111MODULE=on

all: test build

test: generate-code test-lint test-go

test-go:
	go test -v ./...

test-lint:
	golangci-lint run ./...

generate-code:
	./plugin/generateBuiltins.sh $(GOPATH)

build:
	go build -o $(BIN_NAME) cmd/kustomize/main.go

install:
	go install $(PWD)/cmd/kustomize

clean:
	go clean
	rm -f $(BIN_NAME)

.PHONY: test build install clean generate-code test-go test-lint
