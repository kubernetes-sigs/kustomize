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

function runFunc {
  local name=$1
  local result="SUCCESS"
  printf "============== begin %s\n" "$name"
  $name
  local code=$?
  rc=$((rc || $code))
  if [ $code -ne 0 ]; then
    result="FAILURE"
  fi
  printf "============== end %s : %s; exit code=%d\n\n\n" "$name" "$result" $code
}

function testGoLangCILint {
  golangci-lint run ./...
}

function testGoTests {
  go test -v ./...

  if [ -z ${TRAVIS+x} ]; then
    echo " "
    echo Not on travis, so running the notravis Go tests
    echo " "
    
    # Requires helm.
    # At the moment not asking travis to install it.
    go test -v sigs.k8s.io/kustomize/v3/pkg/target \
      -run TestChartInflatorPlugin -tags=notravis
    go test -v sigs.k8s.io/kustomize/v3/plugin/someteam.example.com/v1/chartinflator/... \
      -run TestChartInflator -tags=notravis

    # Requires kubeeval.
    # At the moment not asking travis to install it.
    go test -v sigs.k8s.io/kustomize/v3/plugin/someteam.example.com/v1/validator/... \
       -run TestValidatorHappy -tags=notravis
    go test -v sigs.k8s.io/kustomize/v3/plugin/someteam.example.com/v1/validator/... \
       -run TestValidatorUnHappy -tags=notravis
  fi
}

function testExamplesAgainstLatestRelease {
  /bin/rm -f $HOME/go/bin/kustomize
  # Install latest release.
  go get sigs.k8s.io/kustomize/v3/cmd/kustomize
  PATH=$HOME/go/bin:$PATH \
    mdrip --mode test --label testAgainstLatestRelease ./examples

  if [ -z ${TRAVIS+x} ]; then
    echo " "
    echo Not on travis, so running the notravis example tests
    echo " "

    # Requires helm.  At the moment not asking travis to install it.
    PATH=$HOME/go/bin:$PATH \
      mdrip --mode test --label helmtest README.md ./examples/chart.md
  fi
}

function testExamplesAgainstHead {
  /bin/rm -f $HOME/go/bin/kustomize
  # Install from head.
  go install sigs.k8s.io/kustomize/v3/cmd/kustomize
  # To test examples of unreleased features, add
  # examples with code blocks annotated with some
  # label _other than_ @testAgainstLatestRelease.
  PATH=$HOME/go/bin:$PATH \
    mdrip --mode test --label testAgainstLatestRelease ./examples
}

function generateCode {
  ./plugin/generateBuiltins.sh $oldGoPath
}

# Use of GOPATH is optional if go modules are
# used.  This script tries to work for people who
# don't have GOPATH set, and work for travis.
#
# Upon entry, travis has GOPATH set, and used it
# to install mdrip and the like.
#
# Use GOPATH to define XDG_CONFIG_HOME, then unset
# GOPATH so that go.mod is unambiguously honored.
echo "GOPATH=$GOPATH"

if [ -z ${GOPATH+x} ]; then
  echo GOPATH is unset
  tmp=$HOME/gopath
  if [ -d "$tmp" ]; then
    oldGoPath=$tmp
  else
    tmp=$HOME/go
    if [ -d "$tmp" ]; then
      oldGoPath=$tmp
    fi
  fi
else
  oldGoPath=$GOPATH
  unset GOPATH
fi
echo "oldGoPath=$oldGoPath"
export XDG_CONFIG_HOME=$oldGoPath/src/sigs.k8s.io
echo "XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
if [ ! -d "$XDG_CONFIG_HOME" ]; then
  echo "$XDG_CONFIG_HOME is not a directory."
	exit 1
fi

# Until go v1.13, set this explicitly.
export GO111MODULE=on

echo "HOME=$HOME"
echo "GOPATH=$GOPATH"
echo "GO111MODULE=$GO111MODULE"
echo pwd=`pwd`
echo " "
echo "Working..."

runFunc generateCode
runFunc testGoLangCILint
runFunc testGoTests
runFunc testExamplesAgainstLatestRelease
runFunc testExamplesAgainstHead

if [ $rc -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rc"
fi

exit $rc
