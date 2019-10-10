#!/bin/bash
set -e

# Tracks success or failure of various operations.
# 0==success, any other value is a failure.
rcAccumulator=0

# Not used, and not cross platform,
# but kept because I don't want to have to
# look it up again.
function installHelm {
  wget https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz
  tar -xvzf helm-v2.13.1-linux-amd64.tar.gz
  sudo mv linux-amd64/helm /usr/local/bin/helm
}

function installTools {
  go install sigs.k8s.io/kustomize/pluginator
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

function testGoLangCILint {
  go run "github.com/golangci/golangci-lint/cmd/golangci-lint" run ./...
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
  /bin/rm -f $(go env GOPATH)/bin/kustomize
  # Install latest release.
  (cd ~; go get sigs.k8s.io/kustomize/v3/cmd/kustomize@v3.2.0)

  go run "github.com/monopole/mdrip" --mode test --label testAgainstLatestRelease ./examples

  if [ -z ${TRAVIS+x} ]; then
    echo " "
    echo Not on travis, so running the notravis example tests
    echo " "

    # The following requires helm.
    # At the moment not asking travis to install it.
    go run "github.com/monopole/mdrip" --mode test --label helmtest README.md ./examples/chart.md
  fi
}

function testExamplesAgainstHead {
  /bin/rm -f $(go env GOPATH)/bin/kustomize
  # Install from head.
  (cd kustomize; go install .)
  # To test examples of unreleased features, add
  # examples with code blocks annotated with some
  # label _other than_ @testAgainstLatestRelease.
  go run "github.com/monopole/mdrip" --mode test --label testAgainstLatestRelease ./examples
}

function generateCode {
  ./plugin/generateBuiltins.sh $preferredGoPath
}

# This script tries to work for both travis
# and contributors who have or do not have
# GOPATH set.
#
# Use GOPATH to define XDG_CONFIG_HOME, then unset
# GOPATH so that go.mod is unambiguously honored.

function setPreferredGoPathAndUnsetGoPath {
  preferredGoPath=$GOPATH
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
  echo "preferredGoPath=$preferredGoPath"
}

# Until go v1.13, set this explicitly.
export GO111MODULE=on

# We don't want GOPATH to be defined, as it
# has too much baggage.
setPreferredGoPathAndUnsetGoPath

# This is needed for plugins.
export XDG_CONFIG_HOME=$preferredGoPath/src/sigs.k8s.io
echo "XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
if [ ! -d "$XDG_CONFIG_HOME" ]; then
  echo "$XDG_CONFIG_HOME is not a directory."
  echo "Unable to compile or otherwise work with kustomize plugins."
  exit 1
fi

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
runFunc generateCode
runFunc testGoLangCILint
runFunc testGoTests
runFunc testExamplesAgainstLatestRelease
runFunc testExamplesAgainstHead

if [ $rcAccumulator -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rcAccumulator"
fi

exit $rcAccumulator
