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

function testExamplesAgainstLatestKustomizeRelease {
  removeBin kustomize

  local latest=sigs.k8s.io/kustomize/kustomize/v3
  echo "Installing latest kustomize from $latest"
  (cd ~; GO111MODULE=on go install $latest)

  $MDRIP --mode test \
      --label testAgainstLatestRelease examples

  # TODO: make work for non-linux
  if [[ (-z ${TRAVIS+x}) && ("linux" == "$(go env GOOS)") ]]; then
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

# This script tries to work for both travis
# and contributors who have or do not have
# GOPATH set.
#
# Use GOPATH to define XDG_CONFIG_HOME, then unset
# GOPATH so that go.mod is unambiguously honored.
#
function unsetGoPath {
  local preferredGoPath=$GOPATH
  if [ -z ${GOPATH+x} ]; then
    # GOPATH is unset
    local tmp=$HOME/gopath
    if [ -d "$tmp" ]; then
      preferredGoPath=$tmp
    else
      # this works even if GOPATH undefined.
      preferredGoPath=$(go env GOPATH)
    fi
  else
    unset GOPATH
  fi

  if [ -z ${GOPATH+x} ]; then
    echo GOPATH is unset
  else
    echo "GOPATH=$GOPATH, but should be unset at this point."
    exit 1
  fi

  # This is needed for plugins.
  # TODO: switch this to set KUSTOMIZE_PLUGIN_HOME instead.
  export XDG_CONFIG_HOME=$preferredGoPath/src/sigs.k8s.io
  echo "XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
  if [ ! -d "$XDG_CONFIG_HOME" ]; then
    echo "$XDG_CONFIG_HOME is not a directory."
    echo "Unable to compile or otherwise work with kustomize plugins."
    exit 1
  fi
}

unsetGoPath


# With GOPATH now undefined, this most
# likely this puts $HOME/go/bin on the path.
# Regardless, subsequent go tool installs will
# be placing binaries in this location.
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
