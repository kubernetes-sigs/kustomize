#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e

# run all tests for kyaml and related commands

targets="
kyaml
cmd/config
functions/examples/injection-tshirt-sizes
functions/examples/template-go-nginx
functions/examples/template-heredoc-cockroachdb
functions/examples/validator-kubeval
functions/examples/validator-resource-requests
functions/examples/application-cr
"

# don't run e2e tests in CI by default
export KUSTOMIZE_DOCKER_E2E=${KUSTOMIZE_DOCKER_E2E:-"false"}

for target in $targets; do
  echo "----- Making $target -----"
  pushd .
  cd $target
  make all
  popd
done

# Need better check.  This is repeated git diff check
# more pain than benefit for most people 25Apr2020
## make sure no files were generated or changed by make
## ignore changes to go.mod and go.sum -- they are too flaky
# find . -name go.mod | xargs git checkout --
# find . -name go.sum | xargs git checkout --
# git add .
# git diff-index HEAD --exit-code
