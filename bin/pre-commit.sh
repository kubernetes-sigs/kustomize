#!/bin/bash
set -e

# Make sure, we run in the root of the repo and
# therefore run the tests on all packages
base_dir="$( cd "$(dirname "$0")/.." && pwd )"
cd "$base_dir" || {
  echo "Cannot cd to '$base_dir'. Aborting." >&2
  exit 1
}

rc=0

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

function testGoLangCILint {
  golangci-lint run ./...
}

function testGoTest {
  go test -v ./...
}

function testExamples {
  mdrip --mode test --label test README.md ./examples
}

runTest testGoLangCILint
runTest testGoTest
runTest testExamples

if [ $rc -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rc"
fi

exit $rc
