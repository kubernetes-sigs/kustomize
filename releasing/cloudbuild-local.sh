#!/bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

#
# To test the release process, this script attempts
# to use a Google cloudbuild configuration to create
# release artifacts locally.
#
# See https://cloud.google.com/cloud-build/docs/build-debug-locally
#
# Usage: from the repo root, enter:
#
#   ./releasing/cloudbuild-local.sh kustomize/v1.2.3
#
# or some other valid tag value.
#
# IMPORTANT:
#   The process clones the repo at the given tag,
#   so the repo must have the tag applied upstream.
#   Either use an old tag, or disable the cloud build
#   trigger so that a new testing tag can be applied
#   without setting off a cloud build.

set -e

config=$(mktemp)
cp releasing/cloudbuild.yaml $config

# Add the --snapshot flag to suppress the
# github release and leave the build output
# in the kustomize/dist directory.
sed -i "" "s|# - '--snapshot|- '--snapshot|" $config

echo "Executing cloud-build-local with config file $config :"
echo "========================="
cat $config
echo "========================="

workspace=~/cloud-build-local-workspace

cloud-build-local \
    --config=$config \
    --substitutions=TAG_NAME=$1 \
    --write-workspace=$workspace \
    --dryrun=false \
    .

#  --bind-mount-source \

echo " "
echo "Result of local build:"
echo "##########################################"
tree ./$module/dist
echo "##########################################"
tree ./$workspace
echo "##########################################"
