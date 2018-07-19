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


function testGoCyclo {
  diff <(echo -n) <(go_dirs | xargs -0 gocyclo -over 15)
}

function testGoLint {
  diff -u <(echo -n) <(go_dirs | xargs -0 golint --min_confidence 0.85 )
}

function testGoMetalinter {
  diff -u <(echo -n) <(go_dirs | xargs -0 gometalinter.v2 --disable-all --deadline 5m \
  --enable=misspell \
  --enable=structcheck \
  --enable=deadcode \
# Disabling 'goimports' because it reports hyphens in imported package \
# names as errors, and we have to vendor them in regardless. \
#  --enable=goimports \
  --enable=varcheck \
  --enable=goconst \
  --enable=unparam \
  --enable=ineffassign \
  --enable=nakedret \
  --enable=interfacer \
  --enable=misspell \
  --line-length=170 --enable=lll \
  --dupl-threshold=400 --enable=dupl)
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
runTest testGoMetalinter
runTest testGoLint
runTest testGoVet
runTest testGoCyclo
runTest testGoTest
runTest testExamples

if [ $rc -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rc"
fi

exit $rc
