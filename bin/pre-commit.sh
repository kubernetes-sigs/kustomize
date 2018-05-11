#!/bin/bash

# Make sure, we run in the root of the repo and
# therefore run the tests on all packages
base_dir="$( cd "$(dirname "$0")/.." && pwd )"
cd "$base_dir" || {
  echo "Cannot cd to '$base_dir'. Aborting." >&2
  exit 1
}

rc=0

function go_dirs {
  go list -f '{{.Dir}}' ./... | tr '\n' '\0'
}

function runTest {
  local name=$1
  local result="SUCCESS"
  printf "============== begin %s\n" "$name"
  $name
  local code=$?
  rc=$((rc || $code))
  if [ $code -ne 0 ]; then
    result="FAILURE"
  fi
  printf "============== end %s : %s code=%d\n\n\n" "$name" "$result" $code
}

function testGoFmt {
  diff <(echo -n) <(go_dirs | xargs -0 gofmt -s -d -l)
}

function testGoImports {
  diff -u <(echo -n) <(go_dirs | xargs -0 goimports -l)
}

function testGoVet {
  go vet -all ./...
}

function testGoTest {
  go test -v ./...
}

function testTutorial {
  mdrip --mode test --label test ./cmd/kustomize
}

runTest testGoFmt
runTest testGoImports
runTest testGoVet
runTest testGoTest
runTest testTutorial

exit $rc
