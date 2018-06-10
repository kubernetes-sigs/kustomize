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
  go list -f '{{.Dir}}' ./... | tail -n +2 | tr '\n' '\0'
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

function testGoLint {
  diff -u <(echo -n) <(go_dirs | xargs -0 golint --min_confidence 0.85 )
}

function testGoVet {
  go vet -all ./...
}

function testGoTest {
  go test -v ./...
}

function testExamples {
  mdrip --mode test --label test README.md ./examples
}

runTest testGoFmt
runTest testGoImports
runTest testGoLint
runTest testGoVet
runTest testGoTest
runTest testExamples

if [ $rc -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rc"
fi

exit $rc
