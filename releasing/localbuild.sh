#!/bin/bash

# Usage - from the repository root, enter
#
#   ./releasing/localbuild.sh (kustomize|pluginator)
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
#   releasing/cloudbuild_(kustomize|pluginator|api).yaml
#
# Inside this yaml file is a reference to the script
#
#   releasing/cloudbuild.sh
#
# which runs goreleaser from the proper directory, with the
# proper config.
#
# The script you are reading now does something
# analogous via docker tricks.

set -e

module=$1
case "$module" in
  api)
  ;;
  kustomize)
  ;;
  pluginator)
  ;;
  *)
    echo "Don't recognize module=$module"
    exit 1
  ;;
esac

config=$(mktemp)
cp releasing/cloudbuild_${module}.yaml $config

# Delete the cloud-builders/git step, which isn't needed
# for a local run.
sed -i '2,3d'  $config

# Add the --snapshot flag to suppress the
# github release and leave the build output
# in the kustomize/dist directory.
sed -i 's|"\]$|", "--snapshot"]|' \
    $config

echo "Executing cloud-build-local with:"
echo "========================="
cat $config
echo "========================="

cloud-build-local \
    --config=$config \
    --bind-mount-source \
    --dryrun=false \
    .

echo " "
echo "Result of local build:"
echo "##########################################"
if [ "$module" == "api" ]; then
  tree ./kustapiversion/dist
else
  tree ./$module/dist
fi
echo "##########################################"
