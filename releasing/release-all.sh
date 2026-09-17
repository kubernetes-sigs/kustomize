#!/usr/bin/env bash
# Copyright 2026 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

do_it=0
base_branch="master"
remote=""
release_branch=""
module_bump=""
kustomize_bump=""
module_version=""
kustomize_version=""
github_repo=""
final_pr_url=""

usage() {
  cat <<'EOF'
Usage:
  releasing/release-all.sh \
    --module-bump patch|minor \
    --kustomize-bump patch|minor \
    [--branch release-vX.Y.Z] \
    [--base master] \
    [--remote upstream|origin] \
    [--do-it]

This script automates the multi-module release flow.

It will:
  1. compute the target module versions from the requested bumps
  2. verify kyaml, cmd/config, and api would land on the same version
  3. use gorepomod release with --release-branch for tagging
  4. add the pin commits needed before the downstream releases
  5. add unpin commits and update LATEST_RELEASE
  6. open a PR from the release branch to main

Without --branch, the script uses release-$KUSTOMIZE_VERSION.
Without --do-it, it prints the commands it would run.
Major releases are not supported by this script; use the manual flow instead.
EOF
}

fail() {
  echo "error: $*" >&2
  exit 1
}

log() {
  echo "[$(basename "$0")] $*"
}

run() {
  log "+ $*"
  if [[ "${do_it}" -eq 1 ]]; then
    "$@"
  fi
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

github_repo_from_remote() {
  local remote_url=$1
  gh repo view "${remote_url}" --json nameWithOwner --jq .nameWithOwner
}

pick_default_remote() {
  if git remote get-url upstream >/dev/null 2>&1; then
    printf 'upstream'
    return
  fi
  if git remote get-url origin >/dev/null 2>&1; then
    printf 'origin'
    return
  fi
  fail "unable to find recognized remote; expected one of: upstream, origin"
}

validate_version() {
  local value=$1
  [[ "${value}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "invalid version: ${value}"
}

validate_bump() {
  case "$1" in
    patch|minor) ;;
    *) fail "invalid bump: $1" ;;
  esac
}

bump_version() {
  local version=$1
  local bump=$2
  validate_version "${version}"
  local raw="${version#v}"
  local major minor patch
  IFS='.' read -r major minor patch <<<"${raw}"
  case "${bump}" in
    patch)
      patch=$((patch + 1))
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
  esac
  printf 'v%s.%s.%s' "${major}" "${minor}" "${patch}"
}

module_local_version() {
  local module=$1
  local value
  value="$(go tool gorepomod list --local | awk -v module="${module}" '$1 == module { value = $2 } END { print value }')"
  [[ -n "${value}" ]] || fail "unable to determine local version for module: ${module}"
  printf '%s' "${value}"
}

ensure_clean_workspace() {
  local status
  status="$(git status --porcelain)"
  [[ -z "${status}" ]] || fail "workspace must be clean before running the release automation"
}

ensure_branch_absent_on_remote() {
  local branch=$1
  if git ls-remote --exit-code --heads "${remote}" "${branch}" >/dev/null 2>&1; then
    fail "remote branch already exists: ${remote}/${branch}"
  fi
}

gorepomod_release() {
  local module=$1
  local bump=$2
  run go tool gorepomod release "${module}" "${bump}" --release-branch "${release_branch}" --doIt --local
}

gorepomod_pin() {
  local module=$1
  local version=$2
  run go tool gorepomod pin "${module}" "${version}" --doIt --local
}

gorepomod_unpin() {
  local module=$1
  run go tool gorepomod unpin "${module}" --doIt --local
}

commit_if_needed() {
  local message=$1
  if git diff --quiet --exit-code; then
    log "no changes to commit for: ${message}"
    return
  fi
  run git commit -am "${message}"
}

run_verify() {
  run make IS_LOCAL=true verify-kustomize-repo
}

update_latest_release() {
  run sed -i.bak "s/LATEST_RELEASE=.*/LATEST_RELEASE=${kustomize_version}/" Makefile
  run rm -f Makefile.bak
  if [[ "${do_it}" -eq 1 ]] && ! grep -q "^LATEST_RELEASE=${kustomize_version}$" Makefile; then
    fail "failed to update LATEST_RELEASE in Makefile"
  fi
}

create_final_pr() {
  local title="Release ${kustomize_version}"
  local body
  body=$(cat <<EOF
Merge the release branch for \`${kustomize_version}\` back to \`${base_branch}\`.

- Tags kyaml, cmd/config, and api at \`${module_version}\` and kustomize at \`${kustomize_version}\`.
- Contains the pin commits used during the release.
- Restores in-repo module replacements (unpin).
- Bumps \`LATEST_RELEASE\` to \`${kustomize_version}\`.
EOF
)
  if [[ "${do_it}" -eq 1 ]]; then
    final_pr_url="$(gh pr create \
      --repo "${github_repo}" \
      --base "${base_branch}" \
      --head "${release_branch}" \
      --title "${title}" \
      --body "${body}")"
    printf '%s\n' "${final_pr_url}"
    return
  fi
  run gh pr create \
    --repo "${github_repo:-OWNER/REPO}" \
    --base "${base_branch}" \
    --head "${release_branch}" \
    --title "${title}" \
    --body "${body}"
}

release_action_url() {
  local tag=$1
  local url=""
  if [[ "${do_it}" -ne 1 ]]; then
    printf 'would wait for release workflow run for %s' "${tag}"
    return
  fi

  log "waiting for release workflow run for ${tag}" >&2
  for _ in {1..30}; do
    url="$(gh run list \
      --repo "${github_repo}" \
      --workflow release.yaml \
      --event push \
      --branch "${tag}" \
      --limit 1 \
      --json url \
      --jq '.[0].url // ""')"
    if [[ -n "${url}" ]]; then
      printf '%s' "${url}"
      return
    fi
    sleep 5
  done

  fail "release workflow run was not found for tag: ${tag}"
}

print_postflight_summary() {
  local repo_url="https://github.com/${github_repo}"
  local kyaml_tag="kyaml/${module_version}"
  local cmd_config_tag="cmd/config/${module_version}"
  local api_tag="api/${module_version}"
  local kustomize_tag="kustomize/${kustomize_version}"

  log "release branch automation completed"
  cat <<EOF

Next steps:
  PR: ${final_pr_url:-${repo_url}/pulls}

  Watch release workflows:
    ${kyaml_tag}: $(release_action_url "${kyaml_tag}")
    ${cmd_config_tag}: $(release_action_url "${cmd_config_tag}")
    ${api_tag}: $(release_action_url "${api_tag}")
    ${kustomize_tag}: $(release_action_url "${kustomize_tag}")

  After the workflows finish, open the releases page and undraft the
  4 drafts (${kyaml_tag}, ${cmd_config_tag}, ${api_tag}, ${kustomize_tag}):
    ${repo_url}/releases
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --branch)
      release_branch="${2:-}"
      shift 2
      ;;
    --module-bump)
      module_bump="${2:-}"
      shift 2
      ;;
    --kustomize-bump)
      kustomize_bump="${2:-}"
      shift 2
      ;;
    --base)
      base_branch="${2:-}"
      shift 2
      ;;
    --remote)
      remote="${2:-}"
      shift 2
      ;;
    --do-it)
      do_it=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

[[ -n "${module_bump}" ]] || fail "--module-bump is required"
[[ -n "${kustomize_bump}" ]] || fail "--kustomize-bump is required"

validate_bump "${module_bump}"
validate_bump "${kustomize_bump}"

require_cmd git
require_cmd go
require_cmd gh

cd "${repo_root}"

if [[ -z "${remote}" ]]; then
  remote="$(pick_default_remote)"
fi

git remote get-url "${remote}" >/dev/null 2>&1 || fail "remote not found: ${remote}"
github_repo="$(github_repo_from_remote "$(git remote get-url "${remote}")")"

ensure_clean_workspace

local_kyaml_version="$(module_local_version kyaml)"
local_cmd_config_version="$(module_local_version cmd/config)"
local_api_version="$(module_local_version api)"
local_kustomize_version="$(module_local_version kustomize)"

target_kyaml_version="$(bump_version "${local_kyaml_version}" "${module_bump}")"
target_cmd_config_version="$(bump_version "${local_cmd_config_version}" "${module_bump}")"
target_api_version="$(bump_version "${local_api_version}" "${module_bump}")"
kustomize_version="$(bump_version "${local_kustomize_version}" "${kustomize_bump}")"

if [[ "${target_kyaml_version}" != "${target_cmd_config_version}" || "${target_kyaml_version}" != "${target_api_version}" ]]; then
  fail "requested module bump does not produce a shared version: kyaml=${target_kyaml_version}, cmd/config=${target_cmd_config_version}, api=${target_api_version}"
fi
module_version="${target_kyaml_version}"

if [[ -z "${release_branch}" ]]; then
  release_branch="release-${kustomize_version}"
fi
if [[ "${do_it}" -eq 1 ]]; then
  ensure_branch_absent_on_remote "${release_branch}"
fi

log "base branch: ${base_branch}"
log "release branch: ${release_branch}"
log "remote: ${remote}"
log "GitHub repo: ${github_repo}"
log "module bump: ${module_bump} -> ${module_version}"
log "kustomize bump: ${kustomize_bump} -> ${kustomize_version}"

run git checkout "${base_branch}"
run git fetch "${remote}" "${base_branch}"
run git merge --ff-only "${remote}/${base_branch}"

run_verify

gorepomod_release kyaml "${module_bump}"

gorepomod_pin kyaml "${module_version}"
commit_if_needed "Update kyaml to ${module_version}"
run git push "${remote}" "${release_branch}"
run_verify

gorepomod_release cmd/config "${module_bump}"

gorepomod_pin cmd/config "${module_version}"
commit_if_needed "Update cmd/config to ${module_version}"
run git push "${remote}" "${release_branch}"
run_verify

gorepomod_release api "${module_bump}"

gorepomod_pin api "${module_version}"
commit_if_needed "Update api to ${module_version}"
run git push "${remote}" "${release_branch}"
run_verify

gorepomod_release kustomize "${kustomize_bump}"

gorepomod_unpin api
gorepomod_unpin cmd/config
gorepomod_unpin kyaml
update_latest_release
commit_if_needed "Back to development mode after ${kustomize_version}"
run git push "${remote}" "${release_branch}"

create_final_pr

print_postflight_summary
