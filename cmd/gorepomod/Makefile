include ../../Makefile-modules.mk

.PHONY: test
test: usage.go
	go test -v ./...

usage.go: README.md
	go generate .
	go tool goimports -w usage.go
