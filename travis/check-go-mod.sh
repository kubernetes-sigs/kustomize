#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

# verify all modules pass validation
for i in $(find . -name go.mod); do
  pushd .
  cd $(dirname $i);
  go list -m -json all > /dev/null
  go mod tidy -v
  popd
done

# Need better check.  This is repeated git diff check
# more pain than benefit for most people 25Apr2020
## verify no changes to go.mods -- these should be part of the PR
# find . -name go.sum | xargs git checkout --
# git add .
# git diff-index HEAD --exit-code
