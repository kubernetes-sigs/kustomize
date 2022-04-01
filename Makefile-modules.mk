MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

export KUSTOMIZE_ROOT=$(shell pwd | sed 's|kustomize/.*|kustomize/|')

include $(KUSTOMIZE_ROOT)/Makefile-tools.mk

.PHONY: lint test fix fmt tidy vet

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  -c $$KUSTOMIZE_ROOT/.golangci.yml \
	  --path-prefix $(shell pwd | sed 's|.*kustomize/||') \
	  run ./...

test:
	go test -v -timeout 45m -cover ./...

fix:
	go fix ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...
