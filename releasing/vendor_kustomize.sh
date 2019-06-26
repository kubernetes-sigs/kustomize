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

set -e
set -x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# vendor_kustomize.sh creates the change in kubernetes repo for vendoring kustomize

function setUpWorkspace {
  KPATH=~/kustomize_vendor
  mkdir $KPATH
  GOPATH=$KPATH
}

function cloneK8s {
  mkdir -p $KPATH/src/k8s.io
  cd $KPATH/src/k8s.io

  git clone git@github.com:kubernetes/kubernetes.git
}

function godepRestore {
  cd $KPATH/src/k8s.io/kubernetes

  # restore dependencies
  hack/run-in-gopath.sh hack/godep-restore.sh
}

function getKustomizeDeps {
  # get Kustomize and Kustomize dependencies
  hack/run-in-gopath.sh godep get sigs.k8s.io/kustomize/pkg/commands
  hack/run-in-gopath.sh godep get github.com/bgentry/go-netrc/netrc
  hack/run-in-gopath.sh godep get github.com/hashicorp/go-cleanhttp
  hack/run-in-gopath.sh godep get github.com/hashicorp/go-getter
  hack/run-in-gopath.sh godep get github.com/hashicorp/go-safetemp
  hack/run-in-gopath.sh godep get github.com/hashicorp/go-version

  # The hashes below passed bin/pre-commit.sh with kustomize HEAD at time of merger.
  DEPS=(
  "hashicorp/go-getter     4bda8fa99001c61db3cad96b421d4c12a81f256d"
  "hashicorp/go-cleanhttp  d5fe4b57a186c716b0e00b8c301cbd9b4182694d"
  "hashicorp/go-safetemp   b1a1dbde6fdc11e3ae79efd9039009e22d4ae240"
  "hashicorp/go-version    270f2f71b1ee587f3b609f00f422b76a6b28f348"
  "bgentry/go-netrc        9fd32a8b3d3d3f9d43c341bfe098430e07609480"
  "mitchellh/go-homedir    58046073cbffe2f25d425fe1331102f55cf719de"
  "mitchellh/go-testing-interface  a61a99592b77c9ba629d254a693acffaeb4b7e28"
  "ulikunitz/xz            v0.5.4"
  )

  function foo {
    cd $KPATH/src/k8s.io/kubernetes/_output/local/go/src/github.com/$1
    git checkout $2
  }
  for i in "${DEPS[@]}"; do
    foo $i
  done
}

function updateK8s {
  # Copy k8sdeps from Kustomize to cli-runtime in staging
  mkdir -p $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize
  cp -r $KPATH/src/k8s.io/kubernetes/_output/local/go/src/sigs.k8s.io/kustomize/k8sdeps \
    $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/k8sdeps

  # Change import path of k8sdeps
  find $KPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/k8sdeps \
    -type f -name "*.go" | \
    xargs sed -i \
    's!sigs.k8s.io/kustomize/k8sdeps!k8s.io/cli-runtime/pkg/kustomize/k8sdeps!'


  # Add kustomize command to kubectl
  cp $DIR/vendor_kustomize.diff $KPATH/vendor_kustomize.diff

  cd $GOPATH/src/k8s.io/kubernetes
  git apply --ignore-space-change --ignore-whitespace $KPATH/vendor_kustomize.diff
}

function godepSave {
  # Save all dependencies into k8s.io/kubernetes/vendor by running
  # hack/godep-save.sh
  hack/run-in-gopath.sh  hack/godep-save.sh
}

function verify {
  # make sure in k8s.io/kubernetes/vendor/sigs.k8s.io/kustomize
  # there is no internal package
  test 0 == $(ls $KPATH/src/k8s.io/kubernetes/vendor/sigs.k8s.io/kustomize | grep “internal” | wc -l)

  # Make sure it compiles.
  test 0 == $(bazel build cmd/kubectl:kubectl)

  # next step, open a PR
  echo "The change for vendoring kustomize is ready in $GOPATH/src/k8s.io/kubernetes.\n Next step, open a PR for it.\n"
}

setUpWorkspace
cloneK8s
godepRestore
getKustomizeDeps
updateK8s
godepSave
verify
