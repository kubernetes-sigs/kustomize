#!/bin/bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

#
# This script is called by Kustomize's release pipeline.
# It needs jq (required for release note construction) and [GitHub CLI](https://cli.github.com/).
#
# To test it locally:
#
#   # Please install jq and GitHub CLI. (e.g. macOS)
#   brew install jq gh
#
#   # Setup GitHub CLI
#   gh auth login
#
#   # Run this script, where $TAG is the tag to "release" (e.g. kyaml/v0.13.4)
#   ./releasing/create-release.sh $TAG
#
#   # Please remove Draft Release created by this script.

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1-}" ]]; then
  echo "Usage: $0 TAG"
  echo "  TAG: the tag to build or release, e.g. api/v1.2.3"
  exit 1
fi

git_tag=$1
echo "release tag: $git_tag"

# Build the release binaries for every OS/arch combination.
# It builds compressed artifacts on $release_dir.
function build_kustomize_binary {
  echo "build kustomize binaries"
  version=$1

  release_dir=$2
  echo "build release artifacts to $release_dir"

  mkdir -p "output"
  # build date in ISO8601 format
  build_date=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  for os in linux darwin windows; do
    arch_list=(amd64 arm64)
    if [ "$os" == "linux" ]; then
      arch_list=(amd64 arm64 s390x ppc64le)
    fi
    for arch in "${arch_list[@]}" ; do
      echo "Building $os-$arch"
    #   CGO_ENABLED=0 GOWORK=off GOOS=$os GOARCH=$arch go build -o output/kustomize -ldflags\
      CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o output/kustomize -ldflags\
        "-s -w\
        -X sigs.k8s.io/kustomize/api/provenance.version=$version\
        -X sigs.k8s.io/kustomize/api/provenance.gitCommit=$(git rev-parse HEAD)\
        -X sigs.k8s.io/kustomize/api/provenance.buildDate=$build_date"\
        kustomize/main.go
      if [ "$os" == "windows" ]; then
        zip -j "${release_dir}/kustomize_${version}_${os}_${arch}.zip" output/kustomize
      else
        tar cvfz "${release_dir}/kustomize_${version}_${os}_${arch}.tar.gz" -C output kustomize
      fi
      rm output/kustomize
    done
    rmdir output
  done
}

function create_release {
  git_tag=$1

  # Take everything before the last slash.
  # This is expected to match $module.
  module=${git_tag%/*}

  # Take everything after the last slash.
  version=${git_tag##*/}

  # Generate the changelog for this release
  # using the last two tags for the module
  changelog_file=$(mktemp)
  ./releasing/compile-changelog.sh "$module" "$git_tag" "$changelog_file"

  additional_release_artifacts_arg=""

  # build `kustomize` binary
  if [[ "$module" == "kustomize" ]]; then
    release_artifact_dir=$(mktemp -d)
    build_kustomize_binary "$version" "$release_artifact_dir"

    # additional_release_artifacts_arg+="$release_artifact_dir/*"
    additional_release_artifacts_arg=("$release_artifact_dir"/*)

    # create github releases
    gh release create "$git_tag" \
      --title "$git_tag"\
      --draft \
      --notes-file "$changelog_file"\
      "${additional_release_artifacts_arg[@]}"

    return
  fi

  # create github releases
  gh release create "$git_tag" \
    --title "$git_tag"\
    --draft \
    --notes-file "$changelog_file"
}


## create release
create_release "$git_tag"
