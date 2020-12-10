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

# TODO: change the label?
# We test against the latest release, and HEAD, and presumably
# any branch using this label, so it should probably get
# a new value.
mdrip --mode test --blockTimeOut 9m \
    --label testAgainstLatestRelease examples

# TODO: make work for non-linux
if onLinuxAndNotOnRemoteCI; then
  echo "On linux, and not on remote CI.  Running expensive tests."

  # Requires helm.
  make $(go env GOPATH)/bin/helm
  mdrip --mode test --label helmtest examples/chart.md
fi

echo "Example tests passed against ${version}."
