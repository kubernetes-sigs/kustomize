#!/usr/bin/env bash
#
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o errexit
set -o pipefail

version=$1

function onLinuxAndNotOnTravis {
  [[ ("linux" == "$(go env GOOS)") && (-z ${TRAVIS+x}) ]] && return
  false
}

# TODO: change the label?
# We test against the latest release, and HEAD, and presumably
# any branch using this label, so it should probably get
# a new value.
mdrip --mode test \
    --label testAgainstLatestRelease examples

# TODO: make work for non-linux
if onLinuxAndNotOnTravis; then
  echo "On linux, and not on travis, so running the notravis example tests."

  # Requires helm.
  make $(go env GOPATH)/bin/helm
  mdrip --mode test \
      --label helmtest examples/chart.md
fi

echo "Example tests passed against ${version}."
