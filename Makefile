KUSTOMIZE_NAME=kustomize
PLUGINATOR_NAME=pluginator

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
	cd pluginator && go build -o $(PLUGINATOR_NAME) .
	cd kustomize && go build -o $(KUSTOMIZE_NAME) ./main.go

install:
	cd pluginator && go install $(PWD)/pluginator
	cd kustomize && go install $(PWD)/kustomize

cover:
	# The plugin directory eludes coverage, and is therefore omitted
	go test ./pkg/... ./k8sdeps/... ./internal/... -coverprofile=$(COVER_FILE) && \
	go tool cover -html=$(COVER_FILE)


clean:
	cd kustomize && go clean && rm -f $(KUSTOMIZE_NAME)
	rm -f $(COVER_FILE)

.PHONY: test build install clean generate-code test-go test-lint cover
