#!/bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

#
# Builds and optionally releases the specified module
#
# Usage (from top of repo):
#
#  releasing/run-goreleaser.sh TAG MODE[build|release] [--snapshot]
#
# Where TAG is in the form
#
#   api/v1.2.3
#   kustomize/v1.2.3
#   cmd/config/v1.2.3
#   ... etc.
#

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1-}" || -z "${2-}" ]]; then
  echo "Usage: $0 TAG MODE [goreleaser flags]"
  echo "  TAG: the tag to build or release, e.g. api/v1.2.3"
  echo "  MODE: build or release"
  exit 1
fi

fullTag=$1
shift
echo "fullTag=$fullTag"
export GORELEASER_CURRENT_TAG=$fullTag

if [[ $1 == "release" || $1 == "build" ]]; then
  mode=$1
  shift
else
  echo "Error: mode must be build or release"
  exit 1
fi

remainingArgs="$@"
echo "Remaining args: $remainingArgs"

# Take everything before the last slash.
# This is expected to match $module.
module=${fullTag%/*}
echo "module=$module"

# Take everything after the last slash.
# This should be something like "v1.2.3".
semVer=${fullTag#$module/}
echo "semVer=$semVer"

# Generate the changelog for this release
# using the last two tags for the module
changeLogFile=$(mktemp)
./releasing/compile-changelog.sh "$module" "$fullTag" "$changeLogFile"
echo
echo "######### Release notes: ##########"
cat "$changeLogFile"
echo "###################################"
echo

# This is probably a directory called /workspace

# Sanity check
echo
echo "############ DEBUG ##############"
echo "pwd = $PWD"
echo "ls -las ."
ls -las .
echo "###################################"
echo

# CD into the module directory.
# This directory expected to contain a main.go, so there's
# no need for extra details in the `build` stanza below.
cd $module

# This is used in goreleaser.yaml
skipBuild=true
if [[ "$module" == "kustomize" || "$module" == "pluginator" ]]; then
  # If releasing a main program, don't skip the build.
  skipBuild=false
fi

goReleaserConfigFile=$(mktemp)

cat <<EOF >$goReleaserConfigFile
project_name: $module

archives:
- name_template: "${module}_${semVer}_{{ .Os }}_{{ .Arch }}"

builds:
- skip: $skipBuild

  ldflags: >
    -s
    -X sigs.k8s.io/kustomize/api/provenance.version={{.Version}}
    -X sigs.k8s.io/kustomize/api/provenance.gitCommit={{.Commit}}
    -X sigs.k8s.io/kustomize/api/provenance.buildDate={{.Date}}

  goos:
  - linux
  - darwin
  - windows

  goarch:
  - amd64
  - arm64
  - s390x
  - ppc64le

checksum:
  name_template: 'checksums.txt'

env:
- CGO_ENABLED=0
- GO111MODULE=on
- GOWORK=off

release:
  github:
    owner: kubernetes-sigs
    name: kustomize
  draft: true

EOF

echo
echo "############# CONFIG ##############"
cat "$goReleaserConfigFile"
echo "###################################"
echo

args=(
  --debug
  --timeout 10m
  --parallelism 7
  --config="$goReleaserConfigFile"
  --rm-dist
  --skip-validate
)
if [[ $mode == "release" ]]; then
  args+=(--release-notes="$changeLogFile")
fi

date
export PATH="/usr/local/bin:$PATH"
set -x
time goreleaser "$mode" "${args[@]}" $remainingArgs
date
