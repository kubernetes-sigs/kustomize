#!/bin/bash
# Copyright 2021 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [ "$#" -ne 1 ]; then
     echo "Usage: $0 \$GOYAML_V3_SHA"
     exit 1
 fi

if [ "$(git branch --show-current)" == "master" ]; then
  echo "You must be on a branch to use this script."
  exit 1
fi

blue=$(tput setaf 4)
normal=$(tput sgr0)

# This should be the version of go-yaml v3 used by kubectl
# In the original fork, this is 496545a6307b2a7d7a710fd516e5e16e8ab62dbc
export GOYAML_SHA=$1
export GOYAML_REF="goyaml-$GOYAML_SHA"

# The PRs we need to cherry-pick onto the above commit
declare -r GO_YAML_PRS=(753 766)

REPO_ROOT=$(git rev-parse --show-toplevel)
declare -r REPO_ROOT
declare -r REBASEMAGIC="${REPO_ROOT}/.git/rebase-apply"

function explain() {
    printf "\n\n%s\n" "${blue}$1${normal}"
}

# cherry-pick REPO PR
function cherry-pick(){
  repo=$1
  pull=$2
  echo "+++ Downloading patch to /tmp/${pull}.patch (in case you need to do this again)"

  curl -o "/tmp/${pull}.patch" -sSL "${repo}/pull/${pull}.patch"
  echo
  echo "+++ About to attempt cherry pick of PR. To reattempt:"
  echo "  $ git am -x -X subtree=kyaml/internal/forked/github.com/go-yaml/yaml -3 /tmp/${pull}.patch"
  echo
  git am -3 "/tmp/${pull}.patch" || {
    conflicts=false
    while unmerged=$(git status --porcelain | grep ^U) && [[ -n ${unmerged} ]] \
      || [[ -e "${REBASEMAGIC}" ]]; do
      conflicts=true # <-- We should have detected conflicts once
      echo
      explain "+++ Conflicts detected:"
      echo
      (git status --porcelain | grep ^U) || echo "!!! None. Did you git am --continue?"
      echo
      explain "+++ Please resolve the conflicts in another window (and remember to 'git add / git am --continue')"
      read -p "+++ Proceed (anything but 'y' aborts the cherry-pick)? [y/n] " -r
      echo
      if ! [[ "${REPLY}" =~ ^[yY]$ ]]; then
        explain "Aborting." >&2
        exit 1
      fi
    done

    if [[ "${conflicts}" != "true" ]]; then
      explain "!!! git am failed, likely because of an in-progress 'git am' or 'git rebase'"
      exit 1
    fi
  }

  # remove the patch file from /tmp
  rm -f "/tmp/${pull}.patch"
}

subtree_commit_flag=""

explain "Removing the fork's tree from git, if it exists. We'll write over this commit in a moment, but \`read-tree\` requires a clean directory."
find  kyaml/internal/forked/github.com/go-yaml/yaml -type f -delete

if [[ $(git diff --exit-code kyaml/internal/forked/github.com/go-yaml/yaml) ]]; then
  git add kyaml/internal/forked/github.com/go-yaml/yaml
  git commit -m "Temporarily remove go-yaml fork"
  subtree_commit_flag="--amend"
fi

explain "Fetching the version of go-yaml used by kubectl. Tag it more explicitly in case of conflicts with commits local to this repo."
git fetch --depth=1 https://github.com/go-yaml/yaml.git "$GOYAML_SHA:$GOYAML_REF"

explain "Inserting the content we just pulled as a subtree of this repository and squash the changes into the last commit."
git read-tree --prefix=kyaml/internal/forked/github.com/go-yaml/yaml/ -u "$GOYAML_REF"
git add kyaml/internal/forked/github.com/go-yaml/yaml
git commit $subtree_commit_flag -m "Internal copy of go-yaml at $GOYAML_SHA"

explain "Subtree creation successful."

explain "Cherry-picking the commits from our go-yaml/yaml PRs"
for pr in "${GO_YAML_PRS[@]}" ; do
  cherry-pick https://github.com/go-yaml/yaml "$pr"
done

explain "Converting module to be internal."
find kyaml/internal/forked/github.com/go-yaml/yaml -name "*.go" -type f | xargs sed -i '' s+"gopkg.in/yaml.v3"+"sigs.k8s.io/kustomize/kyaml/internal/forked/github.com/go-yaml/yaml"+g
rm kyaml/internal/forked/github.com/go-yaml/yaml/go.mod
git commit --all -m "Internalize forked code"

explain "SUCCEEDED."
exit 0
