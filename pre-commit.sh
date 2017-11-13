#!/bin/bash

rc=0

go_dirs() {
  go list -f '{{.Dir}}' ./... | tr '\n' '\0'
}

echo "Running go fmt"
go_dirs | xargs -0 gofmt -s -d -l
rc=$((rc || $?))

echo "Running goimports"
diff -u <(echo -n) <(go_dirs | xargs -0 goimports -l)
rc=$((rc || $?))

echo "Running go vet"
go vet -all ./...
rc=$((rc || $?))

echo "Running go test"
go test -v ./...
rc=$((rc || $?))

exit $rc
