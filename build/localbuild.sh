#!/bin/bash

# Usage
#
#   ./build/localbuild.sh
#
# The script attempts to use cloudbuild configuration
# to create a release "locally".
#
# At the time of writing,
#
#   https://pantheon.corp.google.com/cloud-build/triggers?project=kustomize-199618
#
# has a trigger such that whenever a git tag is
# applied to the kustomize repo, the cloud builder
# reads the repository-relative file
#
#   build/cloudbuild.yaml
#  
# Inside this yaml file is a reference to the script
#
#   build/cloudbuild.sh
#
# The script you are reading now does something
# analogous via docker tricks.

set -e
# set -x

if [ -z ${GOPATH+x} ]; then
  echo GOPATH is unset; cannot proceed.
  exit 1
fi

WORK=$(mktemp -d)

pushd $GOPATH/src/sigs.k8s.io/kustomize
pwd

echo "Building in $WORK"

cat <<EOF >/tmp/localbuild.yaml
steps:
- name: "gcr.io/kustomize-199618/golang_with_goreleaser:1.10-stretch"
  args: ["bash", "build/cloudbuild.sh", "--snapshot"]
  secretEnv: ['GITHUB_TOKEN']
secrets:
- kmsKeyName: projects/kustomize-199618/locations/global/keyRings/github-tokens/cryptoKeys/gh-release-token
  secretEnv:
   GITHUB_TOKEN: CiQAyrREbPgXJOeT7M3t+WlxkhXwlMPudixBeiyWTjmLOMLqdK4SUQA0W+xUmDJKAhyfHCcwqSEzUn9OwKC7XAYcmwe0CCKTCbPbDgmioDK24q3LVapndXNvnnHvCjhOJNEr1o+P1DCF+LlzYV2YL8lP09rrKrslPg==
EOF

#  --substitutions=_GOOS=linux,_GOARCH=amd64

config=build/cloudbuild.yaml
config=/tmp/localbuild.yaml

# See https://cloud.google.com/cloud-build/docs/build-debug-locally
cloud-build-local \
  --config=$config \
  --dryrun=false \
  --write-workspace=$WORK \
  .

tree $WORK

popd
