#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script is run periodically by kubernetes
# test-infra.
#
# It uses kustomized configurations in a live cluster,
# to assure that the generated configs work as
# expected.
#
# This script assumes that the process running it has
# checked out the kubernetes-sigs/kustomize repo, and
# has cd'ed into it (i.e. the directory above "examples")
# before running it.
#
# At time of writing, its 'call point' was in
# https://github.com/kubernetes/test-infra/blob/master/config/jobs/kubernetes-sigs/kustomize/kustomize-config.yaml

function exitWith {
  local msg=$1
  echo >&2 ${msg}
  exit 1
}
export -f exitWith

function expectCommand {
  command -v $1 >/dev/null 2>&1 || \
    { exitWith "Expected $1 on PATH."; }
}

function setUpEnv {
  local repo=$(git rev-parse --show-toplevel)
  cd $repo
  [[ $? -eq 0 ]] || "Failed to cd to $repo"
  echo "pwd is " `pwd`

  local expectedRepo=sigs.k8s.io/kustomize
  if [[ `pwd` != */$expectedRepo ]]; then
    exitWith "Script must be run from $expectedRepo"
  fi

  GO111MODULE=on go install ./cmd/kustomize || \
    { exitWith "Failed to install kustomize."; }

  PATH=$GOPATH/bin:$PATH

  expectCommand kustomize
  expectCommand kubectl
}

function runTest {
  local script=$1
  shift
  local args=$@

  if [ ! -x "$script" ]; then
    exitWith "Unable to run $script"
  fi

  $script "$args"
  [[ $? -eq 0 ]] || exitWith "Failed: $script $args"

  echo "$script passed."
}

setUpEnv

pushd examples
runTest ldap/integration_test.sh ldap/base
popd
