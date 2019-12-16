
.PHONY: generate license fix vet fmt test tidy

GOBIN := $(shell go env GOPATH)/bin

all: generate license fix vet fmt test lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	GOBIN=$(GOBIN) go generate ./...

license:
	(which $(GOBIN)/addlicense || go get github.com/google/addlicense)
	$(GOBIN)/addlicense  -y 2019 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint:
	(which $(GOBIN)/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.21.0)
	$(GOBIN)/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...

docker:
	docker build . -t gcr.io/kustomize-functions/v1alpha1:v0.0.1
	docker push gcr.io/kustomize-functions/v1alpha1:v0.0.1
