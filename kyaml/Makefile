# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test lint tidy openapi

GOPATH := $(shell go env GOPATH)

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

# sed command was added to resolve https://github.com/kubernetes-sigs/kustomize/issues/2767
# it changes the ports merge key in kyaml/openapi/kubernetesapi/swagger.json
# likely related to https://github.com/kubernetes/kubernetes/issues/39188
openapi:
	sed -i 's/"x-kubernetes-patch-merge-key": "containerPort"/"x-kubernetes-patch-merge-key": "name"/g' 'openapi/kubernetesapi/swagger.json'
	(which $(GOPATH)/bin/go-bindata || go get -v github.com/go-bindata/go-bindata)
	$(GOPATH)/bin/go-bindata --pkg kubernetesapi -o openapi/kubernetesapi/swagger.go openapi/kubernetesapi/swagger.json
	$(GOPATH)/bin/go-bindata --pkg kustomizationapi -o openapi/kustomizationapi/swagger.go openapi/kustomizationapi/swagger.json

VERSION="v1.19.1"
schema:
	which $(GOPATH)/bin/kind || GO111MODULE=on go get sigs.k8s.io/kind@v0.9.0
	which $(GOPATH)/bin/kpt || GO111MODULE=on go get -v github.com/GoogleContainerTools/kpt
	cp $(HOME)/.kube/config /tmp/kubeconfig.txt | true
	$(GOPATH)/bin/kind create cluster --image kindest/node:$(VERSION) --name=getopenapidata
	$(GOPATH)/bin/kpt live  fetch-k8s-schema  --pretty-print > openapi/kubernetesapi/swagger.json
	kind delete cluster --name=getopenapidata 
	cp /tmp/kubeconfig.txt $(HOME)/.kube/config | true


