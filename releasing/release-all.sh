#!/usr/bin/env bash
# Copyright 2026 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

readonly base_branch="master"
readonly remote="upstream"
# Modules released at the shared module version, in dependency order.
readonly shared_modules=(kyaml cmd/config api)

do_it=0
resume=0
module_bump=""
kustomize_bump=""
module_version=""
kustomize_version=""
release_branch=""
github_repo=""
final_pr_url=""
# Set once this invocation creates a tag; later tags must then be absent.
created_tag=0

usage() {
  cat <<'EOF'
Usage:
  releasing/release-all.sh \
    --module-bump patch|minor \
    --kustomize-bump patch|minor \
    [--resume --module-version vX.Y.Z --kustomize-version vX.Y.Z] \
    [--do-it]

This script automates the multi-module release flow on a single release
branch named release-$KUSTOMIZE_VERSION:

  1. release kyaml, cmd/config, and api at a shared version, pinning each
     one before releasing the modules that depend on it
  2. release kustomize
  3. unpin the modules, update LATEST_RELEASE, and open a PR to master

Without --do-it, it prints the commands it would run.
Major releases are not supported by this script; use the manual flow instead.

Every step can be re-run safely. To resume an interrupted release, re-run the
original command with --resume and the target versions it printed. Existing
tags are never replaced: a tag is skipped only when it is already on the
release branch.
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

validate_version() {
  [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "invalid version: $1"
}

validate_bump() {
  case "$1" in
    patch|minor) ;;
    *) fail "invalid bump: $1" ;;
  esac
}

bump_version() {
  local version=$1 bump=$2 major minor patch
  validate_version "${version}"
  IFS='.' read -r major minor patch <<<"${version#v}"
  case "${bump}" in
    patch) patch=$((patch + 1)) ;;
    minor) minor=$((minor + 1)); patch=0 ;;
  esac
  printf 'v%s.%s.%s' "${major}" "${minor}" "${patch}"
}

module_local_version() {
  local module=$1 value
  value="$(go tool gorepomod list --local | awk -v module="${module}" '$1 == module { value = $2 } END { print value }')"
  [[ -n "${value}" ]] || fail "unable to determine local version for module: ${module}"
  printf '%s' "${value}"
}

next_version() {
  bump_version "$(module_local_version "$1")" "$2"
}

remote_branch_exists() {
  git ls-remote --exit-code --heads "${remote}" "$1" >/dev/null 2>&1
}

remote_tag_exists() {
  git ls-remote --exit-code --tags "${remote}" "refs/tags/$1" >/dev/null 2>&1
}

ensure_tag_is_on_release_branch() {
  local tag=$1
  # Read-only fetch; needed to verify the tag even in a dry run.
  log "+ git fetch ${remote} ${release_branch} refs/tags/${tag}:refs/tags/${tag}"
  git fetch "${remote}" \
    "${release_branch}:refs/remotes/${remote}/${release_branch}" \
    "refs/tags/${tag}:refs/tags/${tag}"
  git merge-base --is-ancestor "${tag}^{commit}" "${remote}/${release_branch}" \
    || fail "tag ${tag} is not on ${remote}/${release_branch}; stop and investigate"
}

# Tags HEAD of the release branch as <module>/<version> and pushes the tag,
# unless the tag already exists on the release branch (resume).
release_module() {
  local module=$1 bump=$2 version=$3
  local tag="${module}/${version}"

  if remote_tag_exists "${tag}"; then
    [[ "${resume}" -eq 1 ]] || fail "remote tag already exists: ${tag}; use --resume only after verifying the interrupted release"
    [[ "${created_tag}" -eq 0 ]] || fail "tag ${tag} exists although an earlier tag was just created; stop and investigate"
    ensure_tag_is_on_release_branch "${tag}"
    log "resume: ${tag} is already released, skipping"
    return
  fi

  local expected
  expected="$(next_version "${module}" "${bump}")"
  [[ "${expected}" == "${version}" ]] || fail "${module} would release ${expected}, not ${version}"
  # Same guard as gorepomod release: a released module must not use replace.
  # Skipped in a dry run, where the preceding pin commands are not executed.
  if [[ "${do_it}" -eq 1 ]] && go mod edit -json "${module}/go.mod" | grep -q '"Replace"'; then
    fail "${module}/go.mod still has replace directives; pin them first"
  fi

  run git tag -a "${tag}" -m "Release ${tag} on branch ${release_branch}"
  run git push "${remote}" "refs/tags/${tag}"
  created_tag=1
}

commit_and_push() {
  local message=$1
  if git diff --quiet --exit-code; then
    log "no changes to commit for: ${message}"
  else
    run git commit -am "${message}"
  fi
  run git push "${remote}" "${release_branch}"
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
  local body
  body=$(cat <<EOF
Merge the release branch for \`${kustomize_version}\` back to \`${base_branch}\`.

- Tags kyaml, cmd/config, and api at \`${module_version}\` and kustomize at \`${kustomize_version}\`.
- Contains the pin commits used during the release.
- Restores in-repo module replacements (unpin).
- Bumps \`LATEST_RELEASE\` to \`${kustomize_version}\`.
EOF
)
  local args=(--repo "${github_repo}" --base "${base_branch}" --head "${release_branch}")
  if [[ "${do_it}" -ne 1 ]]; then
    run gh pr create "${args[@]}" --title "Release ${kustomize_version}" --body "${body}"
    return
  fi
  final_pr_url="$(gh pr list "${args[@]}" --state open --json url --jq '.[0].url // ""')"
  if [[ -n "${final_pr_url}" ]]; then
    log "final PR already exists: ${final_pr_url}"
    return
  fi
  final_pr_url="$(gh pr create "${args[@]}" --title "Release ${kustomize_version}" --body "${body}")"
  log "created final PR: ${final_pr_url}"
}

release_action_url() {
  local tag=$1 url
  if [[ "${do_it}" -ne 1 ]]; then
    printf 'would wait for release workflow run for %s' "${tag}"
    return
  fi
  log "waiting for release workflow run for ${tag}" >&2
  for _ in {1..30}; do
    url="$(gh run list --repo "${github_repo}" --workflow release.yaml --event push \
      --branch "${tag}" --limit 1 --json url --jq '.[0].url // ""')"
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
  local tags=() tag
  for module in "${shared_modules[@]}"; do
    tags+=("${module}/${module_version}")
  done
  tags+=("kustomize/${kustomize_version}")

  log "release branch automation completed"
  echo
  echo "Next steps:"
  echo "  PR: ${final_pr_url:-${repo_url}/pulls}"
  echo
  echo "  Watch release workflows:"
  for tag in "${tags[@]}"; do
    echo "    ${tag}: $(release_action_url "${tag}")"
  done
  echo
  echo "  After the workflows finish, open the releases page and undraft the"
  echo "  ${#tags[@]} drafts (${tags[*]}):"
  echo "    ${repo_url}/releases"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-bump) module_bump="${2:-}"; shift 2 ;;
    --kustomize-bump) kustomize_bump="${2:-}"; shift 2 ;;
    --module-version) module_version="${2:-}"; shift 2 ;;
    --kustomize-version) kustomize_version="${2:-}"; shift 2 ;;
    --resume) resume=1; shift ;;
    --do-it) do_it=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown argument: $1" ;;
  esac
done

[[ -n "${module_bump}" ]] || fail "--module-bump is required"
[[ -n "${kustomize_bump}" ]] || fail "--kustomize-bump is required"
validate_bump "${module_bump}"
validate_bump "${kustomize_bump}"

if [[ "${resume}" -eq 1 ]]; then
  [[ -n "${module_version}" ]] || fail "--module-version is required with --resume"
  [[ -n "${kustomize_version}" ]] || fail "--kustomize-version is required with --resume"
  validate_version "${module_version}"
  validate_version "${kustomize_version}"
elif [[ -n "${module_version}" || -n "${kustomize_version}" ]]; then
  fail "--module-version and --kustomize-version are only supported with --resume"
fi

require_cmd git
require_cmd go
require_cmd gh

cd "${repo_root}"

git remote get-url "${remote}" >/dev/null 2>&1 || fail "remote not found: ${remote}; release-all.sh requires an upstream remote"
github_repo="$(gh repo view "$(git remote get-url "${remote}")" --json nameWithOwner --jq .nameWithOwner)"
[[ -z "$(git status --porcelain)" ]] || fail "workspace must be clean before running the release automation"

if [[ "${resume}" -eq 0 ]]; then
  module_version="$(next_version kyaml "${module_bump}")"
  for module in "${shared_modules[@]}"; do
    v="$(next_version "${module}" "${module_bump}")"
    [[ "${v}" == "${module_version}" ]] || fail "requested module bump does not produce a shared version: kyaml=${module_version}, ${module}=${v}"
  done
  kustomize_version="$(next_version kustomize "${kustomize_bump}")"
fi
release_branch="release-${kustomize_version}"

log "release branch: ${release_branch}"
log "GitHub repo: ${github_repo}"
log "module version: ${module_version} (${module_bump})"
log "kustomize version: ${kustomize_version} (${kustomize_bump})"
log "resume: ${resume}"
if [[ "${resume}" -eq 0 ]]; then
  log "to resume after a failure, re-run with: --resume --module-version ${module_version} --kustomize-version ${kustomize_version}"
fi

if [[ "${resume}" -eq 1 ]]; then
  remote_branch_exists "${release_branch}" || fail "cannot resume: remote branch does not exist: ${remote}/${release_branch}"
  run git fetch "${remote}" "${release_branch}:refs/remotes/${remote}/${release_branch}"
  run git checkout -B "${release_branch}" "${remote}/${release_branch}"
else
  if [[ "${do_it}" -eq 1 ]] && remote_branch_exists "${release_branch}"; then
    fail "remote branch already exists: ${remote}/${release_branch}; use --resume to continue that release"
  fi
  run git fetch "${remote}" "${base_branch}:refs/remotes/${remote}/${base_branch}"
  run git checkout -B "${release_branch}" "${remote}/${base_branch}"
  run git push "${remote}" "${release_branch}"
fi

run_verify

for module in "${shared_modules[@]}"; do
  release_module "${module}" "${module_bump}" "${module_version}"
  run go tool gorepomod pin "${module}" "${module_version}" --doIt --local
  commit_and_push "Update ${module} to ${module_version}"
  run_verify
done

release_module kustomize "${kustomize_bump}" "${kustomize_version}"

for ((i = ${#shared_modules[@]} - 1; i >= 0; i--)); do
  run go tool gorepomod unpin "${shared_modules[i]}" --doIt --local
done
update_latest_release
commit_and_push "Back to development mode after ${kustomize_version}"

create_final_pr

print_postflight_summary
