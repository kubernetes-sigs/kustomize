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
go list ./... | grep -vF pkg/framework/test | xargs go test -v
rc=$((rc || $?))

echo "Running test framework tests"
./pkg/framework/test/scripts/download-binaries.sh \
  && ./pkg/framework/test/scripts/run-tests.sh
rc=$((rc || $?))

exit $rc
