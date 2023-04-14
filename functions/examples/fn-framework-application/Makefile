export KUSTOMIZE_ROOT ?= $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\1|')
include $(KUSTOMIZE_ROOT)/Makefile-modules.mk

CONTROLLER_GEN_VERSION=v0.11.3

generate: $(MYGOBIN)/controller-gen $(MYGOBIN)/embedmd
	go generate ./...
	embedmd -w README.md

build: generate
	go build -v -o $(MYGOBIN)/app-fn cmd/main.go

$(MYGOBIN)/controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

$(MYGOBIN)/embedmd:
	go install github.com/campoy/embedmd@v1.0.0

.PHONY: example
example: build
	$(MYGOBIN)/app-fn pkg/exampleapp/testdata/success/basic/config.yaml


test: generate
	go test -v -timeout 45m -cover ./...
