#!/usr/bin/env bash
#
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o errexit
set -o pipefail

version=$1

# All hack scripts should run from top level.
. hack/shellHelpers.sh

echo "Installing kustomize ${version}"

# Always rebuild, never assume the installed verion is
# the right one to test.
rm -f $(go env GOPATH)/bin/kustomize
if [ "$version" == "HEAD" ]; then
  (cd kustomize; go install .)
else
  GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@${version}
fi

# TODO: change the label?
# We test against the latest release, and HEAD, and presumably
# any branch using this label, so it should probably get
# a new value.
mdrip --mode test --blockTimeOut 15m \
    --label testAgainstLatestRelease examples

# TODO: make work for non-linux
if onLinuxAndNotOnRemoteCI; then
  echo "On linux, and not on remote CI.  Running expensive tests."

  # Requires helm.
  make $(go env GOPATH)/bin/helm
  mdrip --mode test --label helmtest examples/chart.md
fi

# Force outside logic to rebuild kustomize rather than
# rely on whatever this script just did.  Tests should
# be order independent.
echo "Removing kustomize ${version}"
rm $(go env GOPATH)/bin/kustomize

echo "Example tests passed against ${version}."
