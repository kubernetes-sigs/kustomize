#!/usr/bin/env bash
#
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o errexit
set -o pipefail

mdrip --blockTimeOut 60m0s --mode test \
    --label testE2EAgainstLatestRelease examples/alphaTestExamples

echo "Example e2e tests passed against ."
