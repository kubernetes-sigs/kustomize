KUSTOMIZE_NAME      := kustomize
PLUGINATOR_NAME     := pluginator

BINDIR              := bin
TOOLS_DIR           := internal/tools
TOOLS_BIN_DIR       := $(TOOLS_DIR)/bin

# Binaries.
GOLANGCI_LINT       := $(TOOLS_BIN_DIR)/golangci-lint
MDRIP               := $(TOOLS_BIN_DIR)/mdrip
PLUGINATOR          := $(TOOLS_BIN_DIR)/pluginator

COVER_FILE=coverage.out

export GO111MODULE=on

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Build golangci-lint from tools folder.
	cd $(TOOLS_DIR); go build -o ./bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(MDRIP): $(TOOLS_DIR)/go.mod # Build mdrip from tools folder.
	cd $(TOOLS_DIR); go build -o ./bin/mdrip github.com/monopole/mdrip

$(PLUGINATOR): $(TOOLS_DIR)/go.mod # Build pluginator from tools folder.
	cd $(TOOLS_DIR); go build -o ./bin/pluginator sigs.k8s.io/kustomize/pluginator

## --------------------------------------
## Testing
## --------------------------------------

all: test build

.PHONY: test
test: generate-code test-lint test-go

.PHONY: test-go
test-go:
	go test -v ./...

.PHONY: test-lint
test-lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

.PHONY: cover
cover:
	# The plugin directory eludes coverage, and is therefore omitted
	go test ./pkg/... ./k8sdeps/... ./internal/... -coverprofile=$(COVER_FILE) && \
	go tool cover -html=$(COVER_FILE)


.PHONY: generate-code
generate-code: $(PLUGINATOR)
	./plugin/generateBuiltins.sh $(GOPATH)

## --------------------------------------
## Binaries
## --------------------------------------

.PHONY: build
build:
	cd pluginator && go build -o $(PLUGINATOR_NAME) .
	cd kustomize && go build -o $(KUSTOMIZE_NAME) ./main.go

.PHONY: install
install:
	cd pluginator && go install $(PWD)/pluginator
	cd kustomize && go install $(PWD)/kustomize

.PHONY: clean
clean:
	cd kustomize && go clean && rm -f $(KUSTOMIZE_NAME)
	rm -f $(COVER_FILE)
