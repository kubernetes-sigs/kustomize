#!/usr/bin/env bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

ORIGIN_MASTER="feat/5449-add-release-automation"
UPSTREAM_REPO="upstream"
UPSTREAM_MASTER="master"

function createBranch {
  branch=$1
  title=$2
  echo "Making branch $branch : \"$title\""
  # Check if release branch exists
  if git show-ref --quiet "refs/heads/${branch}"; then
    git fetch --tags upstream master
    git checkout $ORIGIN_MASTER
    git branch -D $branch  # delete if it exists
  fi
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
