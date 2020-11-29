# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN := $(shell go env GOPATH)/bin
export PATH := $(MYGOBIN):$(PATH)

.PHONY: generate license fix vet fmt test lint tidy clean
all: license fix vet fmt test lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	(which $(MYGOBIN)/stringer || go get golang.org/x/tools/cmd/stringer)
	go generate ./...

license:
	(which $(MYGOBIN)/addlicense || go get github.com/google/addlicense)
	$(MYGOBIN)/addlicense -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint:
	(which $(MYGOBIN)/golangci-lint || \
	  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	$(MYGOBIN)/golangci-lint \
	  --skip-dirs yaml/internal/k8sgen/pkg \
	  run ./...

test:
	go test -cover ./...

vet:
	go vet ./...

clean:
	rm -rf yaml/internal/k8sgen/pkg
