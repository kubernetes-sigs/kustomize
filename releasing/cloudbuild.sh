#!/bin/bash
#
# This script is called by Kustomize's Cloud Build release pipeline.
# It installs jq (required for release note construction)
# and then runs goreleaser (http://goreleaser.com).
#
# To test it locally, run it in a goreleaser container:
#
#   # Get goreleaser image from cloudbuild.yaml
#   export GORELEASER_IMAGE=goreleaser/goreleaser:v0.179.0
#
#   # Drop into a shell
#   docker run -it --entrypoint=/bin/bash  -v $(pwd):/go/src/github.com/kubernetes-sigs/kustomize -w /go/src/github.com/kubernetes-sigs/kustomize $GORELEASER_IMAGE
#
#   # Run this script in the container, where $TAG is the tag to "release" (e.g. kyaml/v0.13.4)
#   ./releasing/cloudbuild.sh $TAG --snapshot
#

set -e
set -x

fullTag=$1
shift

if ! command -v jq &> /dev/null
then
    # This assumes we are in an alpine container (which is the case for goreleaser images)
    echo "Installing jq."
    apk add jq --no-cache
fi

./releasing/run-goreleaser.sh "$fullTag" release "$@"
