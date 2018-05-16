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
# has cd'ed into it (i.e. the directory above "demos")
# before running it.
#
# At time of writing, its 'call point' was in
# https://github.com/kubernetes/test-infra/blob/master/jobs/config.json

function exit_with {
  local msg=$1
  echo >&2 ${msg}
  exit 1
}
export -f exit_with

repo=kubernetes-sigs/kustomize
if [[ `pwd` != */$repo ]]; then
  exit_with "Script must be run from $repo"
fi

echo    pwd is `pwd`
echo GOPATH is $GOPATH
echo   PATH is $PATH

go install . || \
  { exit_with "Failed to install kustomize."; }

export PATH=$GOPATH/bin:$PATH

command -v kustomize >/dev/null 2>&1 || \
  { exit_with "Require kustomize but it's not installed."; }

command -v kubectl >/dev/null 2>&1 || \
  { exit_with "Require kubectl but it's not installed."; }

function runTest {
  local script=$1
  shift
  local args=$@

  if [ ! -x "$script" ]; then
    exit_with "Unable to run $script"
  fi

  $script "$args"
  if [ $? -ne 0 ]; then
    exit_with "Failed: $script $args"
  fi
  echo "$script passed."
}

pushd demos
runTest ldap/integration_test.sh ldap/base
popd
