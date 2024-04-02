export KUSTOMIZE_ROOT ?= $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\1|')
include $(KUSTOMIZE_ROOT)/Makefile-modules.mk

generate: $(MYGOBIN)/controller-gen $(MYGOBIN)/embedmd
	go generate ./...
	embedmd -w README.md

build: generate
	go build -v -o $(MYGOBIN)/app-fn cmd/main.go

.PHONY: example
example: build
	$(MYGOBIN)/app-fn pkg/exampleapp/testdata/success/basic/config.yaml

test: generate
	go test -v -timeout 45m -cover ./...
