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
  # Copy k8sdeps from Kustomize to cli-runtime in staging
  mkdir -p $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize
  cp -r $GOPATH/src/sigs.k8s.io/kustomize/k8sdeps \
    $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/k8sdeps

  # Change import path of k8sdeps
  find $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io/cli-runtime/pkg/kustomize/k8sdeps \
    -type f -name "*.go" | \
    xargs sed -i \
    's!sigs.k8s.io/kustomize/k8sdeps!k8s.io/cli-runtime/pkg/kustomize/k8sdeps!'


  # Add kustomize command to kubectl
  cat > $GOPATH/visitor.diff << EOF
diff --git a/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/resource/visitor.go b/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/resource/visitor.go
index 32c1a691a5..d7a37e1cde 100644
--- a/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/resource/visitor.go
+++ b/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/resource/visitor.go
@@ -20,10 +20,12 @@ import (
 	"bytes"
 	"fmt"
 	"io"
+	"io/ioutil"
 	"net/http"
 	"net/url"
 	"os"
 	"path/filepath"
+	"strings"
 	"time"

 	"golang.org/x/text/encoding/unicode"
@@ -38,6 +40,9 @@ import (
 	utilerrors "k8s.io/apimachinery/pkg/util/errors"
 	"k8s.io/apimachinery/pkg/util/yaml"
 	"k8s.io/apimachinery/pkg/watch"
+	"k8s.io/cli-runtime/pkg/kustomize/k8sdeps"
+	"sigs.k8s.io/kustomize/pkg/commands/build"
+	"sigs.k8s.io/kustomize/pkg/fs"
 )

 const (
@@ -452,7 +457,10 @@ func ExpandPathsToFileVisitors(mapper *mapper, paths string, recursive bool, ext
 		if err != nil {
 			return err
 		}
-
+		if isKustomizationDir(path) {
+			visitors = append(visitors, NewKustomizationVisitor(mapper, path, schema))
+			return filepath.SkipDir
+		}
 		if fi.IsDir() {
 			if path != paths && !recursive {
 				return filepath.SkipDir
@@ -463,7 +471,10 @@ func ExpandPathsToFileVisitors(mapper *mapper, paths string, recursive bool, ext
 		if path != paths && ignoreFile(path, extensions) {
 			return nil
 		}
-
+		if strings.HasSuffix(path, "kustomization.yaml") {
+			visitors = append(visitors, NewKustomizationVisitor(mapper, filepath.Dir(path), schema))
+			return nil
+		}
 		visitor := &FileVisitor{
 			Path:          path,
 			StreamVisitor: NewStreamVisitor(nil, mapper, path, schema),
@@ -479,6 +490,14 @@ func ExpandPathsToFileVisitors(mapper *mapper, paths string, recursive bool, ext
 	return visitors, nil
 }

+func isKustomizationDir(path string) bool {
+	if _, err := os.Stat(filepath.Join(path, "kustomization.yaml")); err == nil {
+		return true
+	}
+	return false
+}
+
+
 // FileVisitor is wrapping around a StreamVisitor, to handle open/close files
 type FileVisitor struct {
 	Path string
@@ -507,6 +526,37 @@ func (v *FileVisitor) Visit(fn VisitorFunc) error {
 	return v.StreamVisitor.Visit(fn)
 }

+// KustomizationVisitor prorvides the output of kustomization build
+type KustomizationVisitor struct {
+	Path string
+	*StreamVisitor
+}
+
+// Visit in a KustomizationVisitor build the kustomization output
+func (v *KustomizationVisitor) Visit(fn VisitorFunc) error {
+	fSys := fs.MakeRealFS()
+	f := k8sdeps.NewFactory()
+	var out bytes.Buffer
+	cmd := build.NewCmdBuild(&out, fSys, f.ResmapF, f.TransformerF)
+	cmd.SetArgs([]string{v.Path})
+	// we want to silence usage, error output, and any future output from cobra
+	// we will get error output as a golang error from execute
+	cmd.SetOutput(ioutil.Discard)
+	_, err := cmd.ExecuteC()
+	if err != nil {
+		return err
+	}
+	v.StreamVisitor.Reader = bytes.NewReader(out.Bytes())
+	return v.StreamVisitor.Visit(fn)
+}
+
+func NewKustomizationVisitor(mapper *mapper, path string, schema ContentValidator) *KustomizationVisitor {
+	return &KustomizationVisitor{
+		Path:          path,
+		StreamVisitor: NewStreamVisitor(nil, mapper, path, schema),
+	}
+}
+
 // StreamVisitor reads objects from an io.Reader and walks them. A stream visitor can only be
 // visited once.
 // TODO: depends on objects being in JSON format before being passed to decode - need to implement
EOF

  cd $GOPATH/src/k8s.io/kubernetes
  git apply --ignore-space-change --ignore-whitespace $GOPATH/visitor.diff
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
