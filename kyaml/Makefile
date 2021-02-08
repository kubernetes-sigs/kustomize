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

$(MYGOBIN)/addlicense:
	go get github.com/google/addlicense

$(MYGOBIN)/golangci-lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1

$(MYGOBIN)/k8scopy:
	( cd ../cmd/k8scopy; go install . )

$(MYGOBIN)/stringer:
	go get golang.org/x/tools/cmd/stringer

generate: $(MYGOBIN)/stringer $(MYGOBIN)/k8scopy
	go generate ./...

clean:
	rm -rf yaml/internal/k8sgen/pkg

license: $(MYGOBIN)/addlicense
	$(MYGOBIN)/addlicense -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  --skip-dirs yaml/internal/k8sgen/pkg \
	  run ./...

test:
	go test -cover ./...

vet:
	go vet ./...
