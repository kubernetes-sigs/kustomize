BIN_NAME=kustomize

COVER_FILE=coverage.out

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

cover:
	# The plugin directory eludes coverage, and is therefore omitted
	go test ./pkg/... ./k8sdeps/... ./internal/... -coverprofile=$(COVER_FILE) && \
	go tool cover -html=$(COVER_FILE)


clean:
	go clean
	rm -f $(BIN_NAME)
	rm -f $(COVER_FILE)

.PHONY: test build install clean generate-code test-go test-lint cover
