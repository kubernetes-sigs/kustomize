#!/bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

#
# This script is called by Kustomize's Cloud Build release pipeline.
# It installs jq (required for release note construction)
# and then runs goreleaser (http://goreleaser.com).
#
# To test it locally, run it in a goreleaser container:
#
#   # Get build image from cloudbuild.yaml
#   export GOLANG_IMAGE=golang:1.19
#
#   # Drop into a shell
#   docker run -it --entrypoint=/bin/bash -v $(pwd):/go/src/github.com/kubernetes-sigs/kustomize -w /go/src/github.com/kubernetes-sigs/kustomize $GOLANG_IMAGE
#
#   # Run this script in the container, where $TAG is the tag to "release" (e.g. kyaml/v0.13.4)
#   ./releasing/cloudbuild.sh $TAG --snapshot
#

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1-}" ]] ; then
  echo "Usage: $0 <fullTag> [--snapshot]"
  echo "Example: $0 kyaml/v0.13.4"
  exit 1
fi

set -x
fullTag=$1
shift

if ! command -v jq &> /dev/null
then
    # This is expecting to be run from Cloud Build, in a Debian-based official golang container
    echo "Installing jq."
    apt-get update && apt-get install -y jq
fi

if ! command -v goreleaser &> /dev/null
then
    echo "Installing goreleaser."
    make --file Makefile-tools.mk "$(go env GOPATH)/bin/goreleaser"
fi

./releasing/run-goreleaser.sh "$fullTag" release "$@"
