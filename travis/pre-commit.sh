#!/bin/bash
set -e

# Tracks success or failure of various operations.
# 0==success, any other value is a failure.
rcAccumulator=0

function removeBin {
  local d=$(go env GOPATH)/bin/$1
  echo "Removing binary $d"
  /bin/rm -f $d
}

function installTools {
  make install-tools
  MDRIP=$(go env GOPATH)/bin/mdrip
  ls -l $(go env GOPATH)/bin
}

function runFunc {
  local name=$1
  local result="SUCCESS"
  printf "============== begin %s\n" "$name"
  $name
  local code=$?
  rcAccumulator=$((rcAccumulator || $code))
  if [ $code -ne 0 ]; then
    result="FAILURE"
  fi
  printf "============== end %s : %s; exit code=%d\n\n\n" "$name" "$result" $code
}

function runLint {
  make lint
}

function runUnitTests {
  make unit-test-all
}

function onLinuxAndNotOnTravis {
  [[ ("linux" == "$(go env GOOS)") && (-z ${TRAVIS+x}) ]] && return
  false
}

function testExamplesAgainstLatestKustomizeRelease {
  removeBin kustomize

  local latest=sigs.k8s.io/kustomize/kustomize/v3
  echo "Installing latest kustomize from $latest"
  (cd ~; GO111MODULE=on go install $latest)

  $MDRIP --mode test \
      --label testAgainstLatestRelease examples

  # TODO: make work for non-linux
	if onLinuxAndNotOnTravis; then
    echo "On linux, and not on travis, so running the notravis example tests."

    # Requires helm.
    make $(go env GOPATH)/bin/helm
    $MDRIP --mode test \
        --label helmtest examples/chart.md
  fi
  echo "Example tests passed against $latest"
}

function testExamplesAgainstLocalHead {
  removeBin kustomize

  echo "Installing kustomize from HEAD"
  (cd kustomize; go install .)

  # To test examples of unreleased features, add
  # examples with code blocks annotated with some
  # label _other than_ @testAgainstLatestRelease.
  $MDRIP --mode test \
      --label testAgainstLatestRelease examples
  echo "Example tests passed against HEAD"
}

# Having this set might contradict the Makefile,
# so assure it's unset.
unset GOPATH

# With GOPATH now undefined, this most likely
# puts $HOME/go/bin on the path. Regardless,
# subsequent go tool installs will be placing
# binaries in this location.
PATH=$(go env GOPATH)/bin:$PATH

# Make sure we run in the root of the repo and
# therefore run the tests on all packages
base_dir="$( cd "$(dirname "$0")/.." && pwd )"
cd "$base_dir" || {
  echo "Cannot cd to '$base_dir'. Aborting." >&2
  exit 1
}

echo "HOME=$HOME"
echo "PATH=$PATH"
echo pwd=`pwd`
echo " "
echo "Working..."

runFunc installTools
runFunc runLint
runFunc runUnitTests
runFunc testExamplesAgainstLatestKustomizeRelease
runFunc testExamplesAgainstLocalHead

if [ $rcAccumulator -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rcAccumulator"
fi

exit $rcAccumulator
