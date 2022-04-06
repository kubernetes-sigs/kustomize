MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif

include ../../Makefile-modules.mk

$(MYGOBIN)/gorepomod: usage.go
	go install .

.PHONY: test
test: $(MYGOBIN)/gorepomod
	go test -v ./...

usage.go: README.md $(MYGOBIN)/goimports
	go generate . \
	$(MYGOBIN)/goimports -w usage.go
