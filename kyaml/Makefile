# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

.PHONY: generate license fix vet fmt test lint tidy clean

$(MYGOBIN)/addlicense:
	go get github.com/google/addlicense

# TODO: Issue #3663
# Update this version of golangci-lint
# Ideally use same version as in {REPO}/hack/go.mod
$(MYGOBIN)/golangci-lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.0

$(MYGOBIN)/k8scopy:
	( cd ../cmd/k8scopy; go install . )

$(MYGOBIN)/stringer:
	go get golang.org/x/tools/cmd/stringer

all: license fix vet fmt test tidy

k8sGenDir := yaml/internal/k8sgen/pkg

generate: $(MYGOBIN)/stringer $(MYGOBIN)/k8scopy
	go generate ./...

clean:
	rm -rf $(k8sGenDir)

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  run ./...


license: $(MYGOBIN)/addlicense
	( find . -type f -not -path "*/internal/forked/github.com/go-yaml*" -exec bash -c "$(MYGOBIN)/addlicense -y 2021 -c 'The Kubernetes Authors.' -f LICENSE_TEMPLATE {}" ";" )

test:
	go test -cover ./...

fix:
	go fix ./...

fmt:
	go fmt $(shell go list ./... | grep -v "/kyaml/internal/forked/github.com/go-yaml/yaml")

tidy:
	go mod tidy

vet:
	go vet $(shell go list ./... | grep -v "/kyaml/internal/forked/github.com/go-yaml/yaml")
