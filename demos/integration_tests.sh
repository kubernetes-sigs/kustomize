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

home=`pwd`

### LDAP TEST ###
demo_dir="./demos/ldap"
if [ ! -d ${demo_dir} ]; then
  exit_with "directory ${demo_dir} doesn't exist"
fi

test_script="${demo_dir}/integration_test.sh"

if [ -x "${test_script}" ]; then
  ${test_script} ${demo_dir}/base
  if [ $? -eq 0 ]; then
    echo "testing ${demo_dir} passed."
  else
    exit_with "testing ${demo_dir} failed."
  fi
else
  exit_with "Unable to run ${test_script}"
fi
#################
