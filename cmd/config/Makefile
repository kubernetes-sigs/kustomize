# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test build tidy clean

GOBIN = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(shell go env GOPATH)/bin
endif

$(GOBIN)/addlicense:
	go get github.com/google/addlicense

$(GOBIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2

$(GOBIN)/k8scopy:
	( cd ../k8scopy; go install . )

$(GOBIN)/mdtogo:
	go get sigs.k8s.io/kustomize/cmd/mdtogo

build:
	go build -v -o $(GOBIN)/kubectl-krm ./kubectl-krm

all: build license fix vet fmt test lint tidy

k8sGenDir := internal/commands/internal/k8sgen/pkg

generate: $(GOBIN)/mdtogo $(GOBIN)/k8scopy
	GOBIN=$(GOBIN) go generate ./...

clean:
	rm -rf $(k8sGenDir)

lint: $(GOBIN)/golangci-lint
	$(GOBIN)/golangci-lint \
	  --skip-dirs $(k8sGenDir) \
	  run ./... \
	  --path-prefix=cmd/config \
	  -c ../../.golangci.yml

license: $(GOBIN)/addlicense
	$(GOBIN)/addlicense \
	  -y 2022 \
	  -c "The Kubernetes Authors." \
	  -f LICENSE_TEMPLATE .

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
