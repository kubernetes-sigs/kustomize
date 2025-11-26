#!/usr/bin/env bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


function createBranch {
  branch=$1
  title=$2
  echo "Making branch $branch : \"$title\""
  git branch -D $branch  # delete if it exists
  git checkout -b $branch
  git commit -a -m "$title"
  git push -f origin $branch
}

function createPr {
  gh pr create --title "$title" --body "ALLOW_MODULE_SPAN" --base master
}

function refreshMaster {
  git checkout master
  git fetch upstream
  git rebase upstream/master
}

function testKustomizeRepo {
  make IS_LOCAL=true verify-kustomize-repo >& /tmp/k.txt
  local code=$?
  if [ $code -ne 0 ]; then
    echo "**** FAILURE ******************"
    tail /tmp/k.txt
  else
    echo "LGTM"
  fi
}

function nextVersion() {
  local release="$1";
  local libs_release="$2";

  # Even if the major is bumped for kustomize, libs are bumped for minor.
  if [ "$release" = "major"]; then
    libs_release="minor";
  fi

  kustomize_version=$(go tool gorepomod next "$release" kustomize)
  libs_version=$(go tool gorepomod next "$libs_release" kyaml cmd/config api)

  gh workflow run release.yaml -f kustomize_tag="$kustomize_version" -f libs_version="$libs_version"
}