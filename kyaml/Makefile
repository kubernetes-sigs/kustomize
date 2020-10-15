# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN := $(shell go env GOPATH)/bin
GOPATH := $(shell go env GOPATH)
export PATH := $(MYGOBIN):$(PATH)

.PHONY: generate license fix vet fmt test lint tidy openapi schema
all: generate license fix vet fmt test lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	(which $(GOPATH)/bin/stringer || go get golang.org/x/tools/cmd/stringer)
	go generate ./...

license:
	(which $(GOPATH)/bin/addlicense || go get github.com/google/addlicense)
	$(GOPATH)/bin/addlicense  -y 2019 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint:
	(which $(GOPATH)/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	$(GOPATH)/bin/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...

openapi:
	(which $(GOPATH)/bin/go-bindata || go get -u github.com/go-bindata/go-bindata/...)
	$(GOPATH)/bin/go-bindata --pkg kubernetesapi -o openapi/kubernetesapi/swagger.go openapi/kubernetesapi/swagger.json
	$(GOPATH)/bin/go-bindata --pkg kustomizationapi -o openapi/kustomizationapi/swagger.go openapi/kustomizationapi/swagger.json

$(MYGOBIN)/kind:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		wget -O ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(shell uname)-amd64; \
		chmod +x ./kind; \
		mv ./kind $(MYGOBIN); \
		rm -rf $$d; \
	)

$(MYGOBIN)/kpt:
	../hack/install_kpt.sh 0.34.0 $(MYGOBIN)

API_VERSION="v1.19.1"
schema: $(MYGOBIN)/kind $(MYGOBIN)/kpt
	cp $(HOME)/.kube/config /tmp/kubeconfig.txt | true
	$(MYGOBIN)/kind create cluster --image kindest/node:$(API_VERSION) --name=getopenapidata
	$(MYGOBIN)/kpt live  fetch-k8s-schema  --pretty-print > /tmp/new_swagger.json
	$(MYGOBIN)/kind delete cluster --name=getopenapidata
	cp /tmp/kubeconfig.txt $(HOME)/.kube/config | true
	cp /tmp/new_swagger.json openapi/kubernetesapi/swagger.json


