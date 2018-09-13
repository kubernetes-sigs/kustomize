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

# Google Container Builder automatically checks out all the code under the /workspace directory,
# but we actually want it to under the correct expected package in the GOPATH (/go)
# - Create the directory to host the code that matches the expected GOPATH package locations
# - Use /go as the default GOPATH because this is what the image uses
# - Link our current directory (containing the source code) to the package location in the GOPATH

OWNER="sigs.k8s.io"
REPO="kustomize"

GO_PKG_OWNER=$GOPATH/src/$OWNER
GO_PKG_PATH=$GO_PKG_OWNER/$REPO

mkdir -p $GO_PKG_OWNER
ln -sf $(pwd) $GO_PKG_PATH

# When invoked in container builder, this script runs under /workspace which is
# not under $GOPATH, so we need to `cd` to repo under GOPATH for it to build
cd $GO_PKG_PATH


# NOTE: if snapshot is enabled, release is not published to GitHub and the build
# is available under workspace/dist directory.
SNAPSHOT=""

# parse commandline args copied from the link below
# https://stackoverflow.com/questions/192249/how-do-i-parse-command-line-arguments-in-bash?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --snapshot)
    SNAPSHOT="--snapshot"
    shift # past argument
    ;;
esac
done

/goreleaser release --config=build/goreleaser.yaml --rm-dist --skip-validate ${SNAPSHOT}
