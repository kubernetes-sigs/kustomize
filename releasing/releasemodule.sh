#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# run this script with releasing/releasemodule.sh MODULE
#   -- e.g. releasing/releasemodule.sh cmd/config
# to skip fetch from upstream, run with FETCH=false
#   -- e.g. FETCH=false releasing/releasemodule.sh kyaml
set -e

# perform release for a module
function releaseModule {
  # calculate the branch and tag names
  module=$1
  slash="/"
  module_name=${module/$slash/_}
  name="${module_name}_major"
  major="${!name}"
  name="${module_name}_minor"
  minor="${!name}"
  name="${module_name}_patch"
  patch="${!name}"
  branch="release-${module}-v${major}.${minor}"
  tag="${module}/v${major}.${minor}.${patch}"

  # create a temporary workspace for our work
  wktree=$(mktemp -d /tmp/kustomize-releases-XXXXXX)
  git branch $branch upstream/$branch
  git worktree add $wktree $branch # create a separate worktree for the branch
  pushd .
  cd $wktree/$module # cd into the worktree/module

  # merge master changes into the release branch
  git merge upstream/master

  echo "dir:    $wktree"
  echo "module: $module v$major.$minor.$patch"
  echo "branch: $branch"
  echo "tag:    $tag"

  # clean up replaces in go.mod as needed
  FILE=fixgomod.sh
  if test -f "$FILE"; then
    ./fixgomod.sh

    go mod tidy
    go test ./...
    go mod tidy
    git add .
    git commit -m "update go.mod for release" || echo "no changes made to go.mod"
  fi

  if [ "$NO_DRY_RUN" == "true" ]; then
     git push upstream $branch
     git tag -a $tag -m "Release $tag on branch $branch"
     git push upstream $tag
  else
    printf "\nSkipping push module $module -- run with NO_DRY_RUN=true to push the release.\n\n"
  fi

  # cleanup release artifacts
  popd
  rm -rf $wktree
  git worktree prune
  git branch -D $branch

  echo "$module complete"
}

modules="kyaml api kstatus cmd/config cmd/resource cmd/kubectl pluginator kustomize"

# configure the branch and tag names
module="${1?must provide the module to release as an argument: supported modules [$modules]}"

# verify the module
found=false
for m in $modules; do
  if [ "$m" == "$module" ]; then
    found=true
  fi
done
if [ "$found" != "true" ]; then
  echo "unknown module \"$module\", must be one of: [$modules]"
  exit 1
fi

# get the release versions
source releasing/VERSIONS

FETCH=${FETCH:-"true"}
NO_DRY_RUN=${NO_DRY_RUN:-"false"}

# get the most recent changes
if [ "$FETCH" == "true" ]; then
  git fetch upstream
fi

# release the module
releaseModule $module

if [ "$module" == "kustomize" ]; then
  # TODO: Do this for all modules
  pushd .
  getter=$(mktemp -d /tmp/kustomize-releases-XXXXXX)
  cd $getter
  go get sigs.k8s.io/kustomize/$module/v3
  popd
fi
