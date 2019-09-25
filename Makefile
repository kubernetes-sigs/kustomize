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
	cd kustomize && go build -o $(BIN_NAME) ./main.go

install:
	cd kustomize && go install $(PWD)/kustomize

cover:
	# The plugin directory eludes coverage, and is therefore omitted
	go test ./pkg/... ./k8sdeps/... ./internal/... -coverprofile=$(COVER_FILE) && \
	go tool cover -html=$(COVER_FILE)


clean:
	cd kustomize && go clean && rm -f $(BIN_NAME)
	rm -f $(COVER_FILE)

.PHONY: test build install clean generate-code test-go test-lint cover
