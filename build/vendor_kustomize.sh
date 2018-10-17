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

# vendor_kustomize.sh creates the change in kubernetes repo for vendoring kustomize

function setUpWorkspace {
  KPATH=~/kustomize_vendor
  mkdir $KPATH
  GOPATH=$KPATH
}

function cloneK8s {
  mkdir -p $GOPATH/src/k8s.io
  cd $GOPATH/src/k8s.io

  git clone git@github.com:kubernetes/kubernetes.git
}

function godepRestore {
  cd $GOPATH/src/k8s.io/kubernetes

  # restore dependencies
  hack/godep-restore.sh
}

function getKustomizeDeps {
  # get Kustomize and Kustomize dependencies
  godep get sigs.k8s.io/kustomize/pkg/commands
  godep get github.com/bgentry/go-netrc/netrc
  godep get github.com/hashicorp/go-cleanhttp
  godep get github.com/hashicorp/go-getter
  godep get github.com/hashicorp/go-safetemp
  godep get github.com/hashicorp/go-version

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
    cd $GOPATH/src/github.com/$1
    git checkout $2
  }
  for i in "${DEPS[@]}"; do
    foo $i
  done
}

function updateK8s {
  # Copy k8sdeps from Kustomize to kubectl
  mkdir -p $GOPATH/src/k8s.io/kubernetes/pkg/kubectl/kustomize
  cp -r $GOPATH/src/sigs.k8s.io/kustomize/internal/k8sdeps \
    $GOPATH/src/k8s.io/kubernetes/pkg/kubectl/kustomize/k8sdeps

  # Change import path of k8sdeps
  find $GOPATH/src/k8s.io/kubernetes/pkg/kubectl/kustomize/k8sdeps \
    -type f -name "*.go" | \
    xargs sed -i \
    's!sigs.k8s.io/kustomize/internal/k8sdeps!k8s.io/kubernetes/pkg/kubectl/kustomize/k8sdeps!'


  # Add kustomize command to kubectl
  cat > $GOPATH/kubectl.diff << EOF
diff --git a/pkg/kubectl/cmd/cmd.go b/pkg/kubectl/cmd/cmd.go
index 43a541ecc9..2d23bfd27d 100644
--- a/pkg/kubectl/cmd/cmd.go
+++ b/pkg/kubectl/cmd/cmd.go
@@ -74,6 +74,8 @@ import (
        "k8s.io/kubernetes/pkg/kubectl/util/templates"

        "k8s.io/cli-runtime/pkg/genericclioptions"
+       "k8s.io/kubernetes/pkg/kubectl/kustomize/k8sdeps"
+       "sigs.k8s.io/kustomize/pkg/commands"
 )

 const (
@@ -505,6 +507,7 @@ func NewKubectlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
                                replace.NewCmdReplace(f, ioStreams),
                                wait.NewCmdWait(f, ioStreams),
                                convert.NewCmdConvert(f, ioStreams),
+                               templates.NormalizeAll(commands.NewDefaultCommand(k8sdeps.NewFactory())),
                        },
                },
                {

EOF

  cd $GOPATH/src/k8s.io/kubernetes
  git apply --ignore-space-change --ignore-whitespace $GOPATH/kubectl.diff
}

function godepSave {
  # Save all dependencies into k8s.io/kubernetes/vendor by running
  # hack/godep-save.sh
  ./hack/godep-save.sh
}

function verify {
  # make sure in k8s.io/kubernetes/vendor/sigs.k8s.io/kustomize
  # there is no internal package
  test 0 == $(ls $GOPATH/src/k8s.io/kubernetes/vendor/sigs.k8s.io/kustomize | grep “internal” | wc -l)

  # Make sure it compiles.
  test 0 == $(bazel build cmd/kubectl:kubectl)

  # next step, open a PR
  echo "The change for vendoring kustomize is ready in $GOPATH/src/k8s.io/kubernetes.\n Next step, open a PR for it.\n"
}

function updateDocs {
  ./hack/update-generated-docs.sh
}

setUpWorkspace
cloneK8s
godepRestore
getKustomizeDeps
updateK8s
godepSave
verify
updateDocs
