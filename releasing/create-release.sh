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

declare -a RELEASE_TYPES=("major" "minor" "patch")
upstream_master="master"
origin_master="master"

if [[ -z "${2-}" ]]; then
  echo "Release type not specified, using default value: patch"
  release_type="patch"
elif [[ ! "${RELEASE_TYPES[*]}" =~ "${2}" ]]; then
  echo "Unsupported release type, only input these values: major, minor, patch."
  exit 1
fi

function create_release {

  # Take everything before the last slash.
  # This is expected to match $module.
  module=${git_tag%/*}
  module_slugified=$(echo $module | iconv -t ascii//TRANSLIT | sed -E -e 's/[^[:alnum:]]+/-/g' -e 's/^-+|-+$//g' | tr '[:upper:]' '[:lower:]')

  # Create release branch release-{module}/{version}
  echo "Creating release branch $release_branch..."
  git fetch --tags upstream $upstream_master
  git branch $release_branch $origin_master
  git commit -a -m "create release branch $release_branch" || true
  git push -f origin $release_branch

  # Trigger workflow for respective modeule release
  gh workflow run "release-${module_slugified}.yml" -f "release_type=${release_type}" -f "release_branch=${release_branch}"
}

function determineNextVersion {
    module=$1
    currentTag=$(git tag --list "${module}*"  --sort=-creatordate | head -n1)
    currentVersion=$(echo ${currentTag##*/} | cut -d'v' -f2)
    majorVer=$(echo $currentVersion | cut -d'.' -f1)
    minorVer=$(echo $currentVersion | cut -d'.' -f2)
    patchVer=$(echo $currentVersion | cut -d'.' -f3)

    if [[ ${release_type} == "major" ]]; then
      majorVer="$(($majorVer + 1))"
    elif [[ ${release_type} == "minor" ]]; then
      minorVer="$(($minorVer + 1))"
    elif [[ ${release_type} == "patch" ]]; then
      patchVer="$(($patchVer + 1))"
    else
      echo "Error: release_type not supported. Available values 'major', 'minor', 'patch'"
      exit 1
    fi

    echo "$majorVer.$minorVer.$patchVer"
}

main() {

  module=$1
  release_type=$2
  nextVersion=$(determineNextVersion $module)

  # Release branch naming format: release-{module}-{version}
  release_branch="release-${module}-v${nextVersion}"

  # Release tag naming format: {module}/{version}
  git_tag="${module}/${nextVersion}"

  echo "module: $module"
  echo "release type: $release_type"
  echo "tag: ${git_tag}"
  
  ## create release
  create_release "$git_tag"
}

main $@