#!/bin/bash
#
#  Copyright 2018 The Kubernetes Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# This script does the following steps
#
#  0. make a workspace ~/kustomize_vendor
#  1. clone Kubernetes repo
#  2. clone Kustomize repo
#  3. copy 4 directories in Kustomize reop
#         internal
#         k8sdeps
#         pkg
#         plugin
#     into staging/src/k8s.io/cli-runtime/kustomize
#   4. update the import path
#   5. apply the patch
#   6. update vendor and update bazel files
#   7. verify that kubectl binary can be built
#   8 verify that all tests pass
#
# The script will make 3 commits inside the Kubernetes repo:
#    1. copy a Kustomize snapshot
#    2. update cli-runtime and kubectl
#    3. update vendor and bazel files
#
# Then one can open a PR in the remote kubernetes repo for this change.
#
# Copying a snapshot of Kustomize into kubectl doesn't new dependency(This can be confirmed by
# viewing the change in go.mod.). When moving kubectl out from kubernetes, the
# snapshot of Kustomize doesn't cause extra work.
#

set -e
set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# vendor_kustomize.sh creates the change in kubernetes repo for vendoring kustomize

function setUpWorkspace {
  KPATH=~/kustomize_vendor
  GOPATH=$KPATH
}

function cloneRepos {
  mkdir $KPATH

  mkdir -p $KPATH/src/k8s.io
  cd $KPATH/src/k8s.io
  git clone git@github.com:kubernetes/kubernetes.git

  mkdir -p $KPATH/src/sigs.k8s.io
  cd $KPATH/src/sigs.k8s.io
  git clone git@github.com:kubernetes-sigs/kustomize.git
}

function copyKustomizeSnapShot {
  rm -r $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/k8sdeps
  for dir in k8sdeps internal pkg plugin
  do
    cp -r  $KPATH/src/sigs.k8s.io/kustomize/${dir} $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/${dir}
    changeImportPath ${dir}
  done

  # remove test files
  for dir in k8sdeps internal pkg plugin
  do
    find $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/${dir} -name "*_test.go" | xargs rm
  done

  cd $KPATH/src/k8s.io/kubernetes
  git add .
  test 0 == $(git commit -m 'copy a Kustomize snapshot')
}

function changeImportPath {
  # change import path of kustomize
  find $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/$1 \
      -type f -name "*.go" | \
      xargs sed -i \
      's!sigs.k8s.io/kustomize/!k8s.io/cli-runtime/pkg/kustomize/!'
}

function applyChange {
  # apply changes to cli-runtime and kubectl
  cp $DIR/vendor_kustomize_2.1.0.diff $KPATH/vendor_kustomize.diff

  cd $GOPATH/src/k8s.io/kubernetes
  git apply --ignore-space-change --ignore-whitespace $KPATH/vendor_kustomize.diff

  cd $KPATH/src/k8s.io/kubernetes
  git add .
  git commit -m 'update cli-runtime and kubectl'
}

function updateK8s {
    $KPATH/src/k8s.io/kubernetes/hack/update-vendor.sh
    $KPATH/src/k8s.io/kubernetes/hack/update-bazel.sh
    cd $KPATH/src/k8s.io/kubernetes
    git add .
    git commit -m 'update vendor and bazel files'
}

function verify {
  cd $KPATH/src/k8s.io/kubernetes

  # Make sure it compiles.
  bazel build cmd/kubectl:kubectl
  test 0 == $?

  # Make sure the tests pass
  make test
  test 0 == $?

  # next step, open a PR
  echo "The change for vendoring kustomize is ready in $GOPATH/src/k8s.io/kubernetes.\n Next step, open a PR for it.\n"
}

setUpWorkspace
cloneRepos
copyKustomizeSnapShot
applyChange
updateK8s
verify
