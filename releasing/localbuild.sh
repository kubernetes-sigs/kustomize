#!/bin/bash

# Usage
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
# The script you are reading now does something
# analogous via docker tricks.

set -e

if [ -z ${GOPATH+x} ]; then
  echo GOPATH is unset; cannot proceed.
  exit 1
fi

pushd $GOPATH/src/sigs.k8s.io/kustomize
pwd

# The first "step" in the following uses a special
# goreleaser container image that the kubebuilder folks made.
# TODO: On a rainy day, switch to something more standard.

config=$(mktemp)
cat <<EOF >$config
steps:
- name: "gcr.io/kubebuilder/goreleaser_with_go_1.12.5:0.0.1"
  args: ["bash", "releasing/cloudbuild.sh", "--snapshot"]
  secretEnv: ['GITHUB_TOKEN']
secrets:
- kmsKeyName: projects/kustomize-199618/locations/global/keyRings/github-tokens/cryptoKeys/gh-release-token
  secretEnv:
   GITHUB_TOKEN: CiQAyrREbPgXJOeT7M3t+WlxkhXwlMPudixBeiyWTjmLOMLqdK4SUQA0W+xUmDJKAhyfHCcwqSEzUn9OwKC7XAYcmwe0CCKTCbPbDgmioDK24q3LVapndXNvnnHvCjhOJNEr1o+P1DCF+LlzYV2YL8lP09rrKrslPg==
EOF

cloud-build-local \
  --config=$config \
  --bind-mount-source \
  --dryrun=false \
  .

# Print results of local build, which went to ./dist
echo "##########################################"
tree ./dist
echo "##########################################"

popd
