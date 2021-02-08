# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN := $(shell go env GOPATH)/bin
export PATH := $(MYGOBIN):$(PATH)

.PHONY: generate license fix vet fmt test lint tidy clean

$(MYGOBIN)/addlicense:
	go get github.com/google/addlicense

$(MYGOBIN)/golangci-lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1

$(MYGOBIN)/k8scopy:
	( cd ../cmd/k8scopy; go install . )

$(MYGOBIN)/stringer:
	go get golang.org/x/tools/cmd/stringer

all: license fix vet fmt test lint tidy

k8sGenDir := yaml/internal/k8sgen/pkg

generate: $(MYGOBIN)/stringer $(MYGOBIN)/k8scopy
	go generate ./...

clean:
	rm -rf $(k8sGenDir)

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  --skip-dirs $(k8sGenDir) \
	  run ./...

license: $(MYGOBIN)/addlicense
	$(MYGOBIN)/addlicense \
	  -y 2021 \
	  -c "The Kubernetes Authors." \
	  -f LICENSE_TEMPLATE .

test: generate
	go test -cover ./...

fix:
	go fix ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...
