#!/bin/bash

set -eo pipefail
set -o nounset

declare -a RELEASE_TYPES=("major" "minor" "patch")

if [[ -z "${2-}" ]]; then
  echo "Release type not specified, using default value: patch"
  release_type="patch"
elif [[ ! "${RELEASE_TYPES[*]}" =~ "${2}" ]]; then
  echo "Unsupported release type, only input these values: major, minor, patch."
  exit 1
fi

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
  echo "v$nextVersion"
  exit 0
}

main "$@"
