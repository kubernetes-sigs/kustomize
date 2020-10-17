# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN := $(shell go env GOPATH)/bin

.PHONY: all
all: \
	kustomizationapi/swagger.go \
	kubernetesapi/swagger.go \
	kubernetesapi/openapiinfo.go

.PHONY: clean
clean:
	rm kustomizationapi/swagger.go
	rm kubernetesapi/swagger.go
	rm kubernetesapi/openapiinfo.go

# To get swagger.json, we need a cluster at the correct version,
# so think twice before deleting.
.PHONY: nuke
nuke: clean
	rm kubernetesapi/swagger.json

$(MYGOBIN)/go-bindata:
	go install github.com/go-bindata/go-bindata/v3/go-bindata

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
	../../hack/install_kpt.sh 0.34.0 $(MYGOBIN)

kubernetesapi/swagger.go: $(MYGOBIN)/go-bindata kubernetesapi/swagger.json
	$(MYGOBIN)/go-bindata \
		--pkg kubernetesapi \
		-o kubernetesapi/swagger.go \
		kubernetesapi/swagger.json

kustomizationapi/swagger.go: $(MYGOBIN)/go-bindata kustomizationapi/swagger.json
	$(MYGOBIN)/go-bindata \
		--pkg kustomizationapi \
		-o kustomizationapi/swagger.go \
		kustomizationapi/swagger.json

kubernetesapi/openapiinfo.go: kubernetesapi/swagger.json
	./makeOpenApiInfoDotGo.sh kubernetesapi/swagger.json

API_VERSION="v1.19.1"
kubernetesapi/swagger.json: $(MYGOBIN)/kind $(MYGOBIN)/kpt
	cp $(HOME)/.kube/config /tmp/kubeconfig.txt | true
	$(MYGOBIN)/kind create cluster --image kindest/node:$(API_VERSION) --name=getopenapidata
	$(MYGOBIN)/kpt live  fetch-k8s-schema  --pretty-print > /tmp/new_swagger.json
	$(MYGOBIN)/kind delete cluster --name=getopenapidata
	cp /tmp/kubeconfig.txt $(HOME)/.kube/config | true
	cp /tmp/new_swagger.json kubernetesapi/swagger.json
