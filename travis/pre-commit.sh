#!/bin/bash
set -e

# Tracks success or failure of various operations.
# 0==success, any other value is a failure.
rcAccumulator=0

# Not used here, and not cross platform, but kept because
# I don't want to have to look it up again.
function installHelm {
  local d=$(mktemp -d)
  pushd $d
  wget https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz
  tar -xvzf helm-v2.13.1-linux-amd64.tar.gz
  sudo mv linux-amd64/helm /usr/local/bin/helm
  popd
}

# Not used here, and not cross platform, but kept because
# I don't want to have to look it up again.
# Per https://kubeval.instrumenta.dev/installation
function installKubeval {
  local d=$(mktemp -d)
  pushd $d
  wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-linux-amd64.tar.gz
  tar xf kubeval-linux-amd64.tar.gz
  sudo cp kubeval /usr/local/bin
  popd
}

function removeBin {
  local d=$(go env GOPATH)/bin/$1
  echo "Removing binary $d"
  /bin/rm -f $d
}

function installTools {
  # TODO(2019/Oct/19): After the API is in place, and
  # there's a new pluginator compiled against that API,
  # switch back to this:
  #  go install sigs.k8s.io/kustomize/pluginator
  # In the meantime, use the local copy.
  # Go module rules, and the existing violations of
  # semver, mean that simply using the replace directive
  # in the go.mod won't yield the desired result.

  removeBin pluginator
  # Install from whatever code is on disk.
  (cd pluginator; go install .)
  echo "Installed pluginator."

  # TODO figure out how to express this dependence in the three
  # modules (kustomize/api, kustomize/pluginator, kustomize/kustomize).
  # Maybe make a top level module with an internal/tools/tools.go with
  # import _ "github.com/golangci/golangci-lint/cmd/golangci-lint" etc?
  # but it will be a kustomize module, and that will be confusing
  # to people accustomied to the old one-module scheme.  Will
  # require setting the module to at least v4, because it is
  # incompatible with v3.
  removeBin golangci-lint
  GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1
  CILINT=$(go env GOPATH)/bin/golangci-lint

  removeBin mdrip
  GO111MODULE=on go get github.com/monopole/mdrip@v1.0.0
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

function testGoLangCILint {
  (cd api; $CILINT run ./...)
  (cd kustomize; $CILINT run ./...)
  (cd pluginator; $CILINT run ./...)
}

function runApiModuleGoTests {
  (cd api; go test ./...)

  if [ -z ${TRAVIS+x} ]; then
    echo " "
    echo Not on travis, so running the notravis Go tests
    echo " "

    # Requires helm.
    # At the moment not asking travis to install it.
    (cd api; go test -v sigs.k8s.io/kustomize/api/target \
      -run TestChartInflatorPlugin -tags=notravis)
    (cd plugin/someteam.example.com/v1/chartinflator;
     go test -v . -run TestChartInflator -tags=notravis)

    # Requires kubeeval.
    # At the moment not asking travis to install it.
    (cd plugin/someteam.example.com/v1/validator;
     go test -v . -run TestValidatorHappy -tags=notravis)
    (cd plugin/someteam.example.com/v1/validator;
     go test -v . -run TestValidatorUnHappy -tags=notravis)
  fi
}

function testExamplesAgainstLatestKustomizeRelease {
  removeBin kustomize

  local latest=sigs.k8s.io/kustomize/kustomize/v3
  echo "Installing latest kustomize from $latest"
  (cd ~; GO111MODULE=on go install $latest)

  (cd api;
   $MDRIP --mode test \
     --label testAgainstLatestRelease ../examples)

  if [ -z ${TRAVIS+x} ]; then
    echo "Not on travis, so running the notravis example tests."

    # The following requires helm.
    # At the moment not asking travis to install it.
    (cd api;
     $MDRIP --mode test \
       --label helmtest README.md ../examples/chart.md)
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
  (cd api;
   $MDRIP --mode test \
     --label testAgainstLatestRelease ../examples)
  echo "Example tests passed against HEAD"
}

function generateCode {
  echo "preferredGoPath = $preferredGoPath"
  ./api/plugins/builtinhelpers/generateBuiltins.sh $preferredGoPath
}

# This script tries to work for both travis
# and contributors who have or do not have
# GOPATH set.
#
# Use GOPATH to define XDG_CONFIG_HOME, then unset
# GOPATH so that go.mod is unambiguously honored.
#
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
runFunc runApiModuleGoTests
runFunc testExamplesAgainstLatestKustomizeRelease
runFunc testExamplesAgainstLocalHead

if [ $rcAccumulator -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rcAccumulator"
fi

exit $rcAccumulator
