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


# This script run periodically by kubernetes test-infra.
# At time of writing, it's 'call point' was in
# https://github.com/kubernetes/test-infra/blob/master/jobs/config.json

function exit_with {
  local msg=$1
  echo >&2 ${msg}
  exit 1
}

base_dir="$( cd "$(dirname "$0")/../../.." && pwd )"
cd "$base_dir" || {
  exit_with "Cannot cd to ${base_dir}. Aborting."
}

go install github.com/kubernetes-sigs/kustomize || \
  { exit_with "Failed to install kustomize"; }
export PATH=$GOPATH/bin:$PATH

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

runTest ldap/integration_test.sh ldap/base
