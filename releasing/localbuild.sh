#!/bin/bash

# Usage - from the repository root, enter
#
#   ./releasing/localbuild.sh
#
# The script attempts to use cloudbuild configuration
# to create a release "locally".
#
# See https://cloud.google.com/cloud-build/docs/build-debug-locally
#
# At the time of writing,
#
#   https://pantheon.corp.google.com/cloud-build/triggers?project=kustomize-199618
#
# has a trigger such that whenever a git tag is
# applied to the kustomize repo, the cloud builder
# reads the repository-relative file
#
#   releasing/cloudbuild.yaml
#  
# Inside this yaml file is a reference to the script
#
#   releasing/cloudbuild.sh
#
# which runs goreleaser from the proper directory.
#
# The script you are reading now does something
# analogous via docker tricks.

set -e

# Modify cloudbuild.yaml to add the --snapshot flag.
# This suppresses the github release, and leaves
# the build output in the kustomize/dist directory.
config=$(mktemp)
sed 's|\["releasing/cloudbuild.sh"\]|["releasing/cloudbuild.sh", "--snapshot"]|' \
    releasing/cloudbuild.yaml > $config

cloud-build-local \
    --config=$config \
    --bind-mount-source \
    --dryrun=false \
    .

# Print results of local build
echo "##########################################"
tree ./kustomize/dist
echo "##########################################"
