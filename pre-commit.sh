#!/bin/bash

# Make sure, we run in the root of the repo and
# therefore run the tests on all packages
base_dir="$( cd "$(dirname "$0")/.." && pwd )"
cd "$base_dir" || {
  echo "Cannot cd to '$base_dir'. Aborting." >&2
  exit 1
}

rc=0

go_dirs() {
  go list -f '{{.Dir}}' ./... | tr '\n' '\0'
}

echo "Running go fmt"
diff <(echo -n) <(go_dirs | xargs -0 gofmt -s -d -l)
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

echo "Testing kinflate demos"
mdrip --mode test --label test ./cmd/kinflate
rc=$((rc || $?))

exit $rc
