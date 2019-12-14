#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e

# fetch upstream once
git fetch upstream
export FETCH="false"

# release modules without binaries
for module in "kyaml api cmd/config cmd/kubectl"
do
  releasing/releasemodule.sh $module
done

# release modules with binaries
for binary in "kustomize"
do
  BINARY=true releasing/releasemodule.sh $binary
done
