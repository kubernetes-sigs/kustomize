#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

usage() {
  echo "usage: $0 <go patch version: 1.N.M> [--check]"
}

patch=""
check_only=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)
      check_only=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      if [[ -n "${patch}" ]]; then
        echo "unexpected argument: $1" >&2
        usage >&2
        exit 2
      fi
      patch="$1"
      shift
      ;;
  esac
done

if [[ -z "${patch}" ]]; then
  usage >&2
  exit 2
fi

if [[ ! "${patch}" =~ ^1\.[0-9]+\.[0-9]+$ ]]; then
  echo "patch must look like 1.N.M, e.g. 1.25.10" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd "${script_dir}/.." && pwd -P)"
cd "${repo_root}"

if [[ ! -f go.work ]]; then
  echo "could not find go.work at repository root: ${repo_root}" >&2
  exit 1
fi

for tool in git go awk mktemp cp mv; do
  if ! command -v "${tool}" >/dev/null 2>&1; then
    echo "missing required command: ${tool}" >&2
    exit 1
  fi
done

minor="${patch%.*}"
go_directive="${minor}.0"
toolchain="go${patch}"

echo "go directive       : ${go_directive}"
echo "toolchain / Docker : ${patch}"

module_files() {
  git ls-files -z -- \
    'go.mod' \
    'go.mod.src' \
    ':(glob)**/go.mod' \
    ':(glob)**/go.mod.src'
}

image_files() {
  local status

  git --no-pager grep -zlE 'golang:1\.[0-9]+\.[0-9]+' -- \
    '*Dockerfile' \
    '*.Dockerfile' \
    '*.go' \
    '*.md' \
    '*.yaml' \
    '*.yml' || {
      status=$?
      [[ "${status}" -eq 1 ]] || return "${status}"
    }
}

go_edit() {
  GOWORK=off go "$@"
}

replace_golang_image_versions() {
  local file="$1"
  local tmp

  tmp="$(mktemp "${file}.XXXXXX")"
  cp -p "${file}" "${tmp}"
  if ! awk -v patch="${patch}" '{
    gsub(/golang:1\.[0-9]+\.[0-9]+/, "golang:" patch)
    print
  }' "${file}" >"${tmp}"; then
    rm -f "${tmp}"
    return 1
  fi
  mv "${tmp}" "${file}"
}

check_versions() {
  local actual
  local failed=0
  local found_images=false
  local found_modules=false
  local image_ref
  local modfile

  while IFS= read -r -d '' modfile; do
    found_modules=true
    actual="$(awk '$1 == "go" { print $2; exit }' "${modfile}")"
    if [[ "${actual}" != "${go_directive}" ]]; then
      echo "${modfile}: expected 'go ${go_directive}', found 'go ${actual:-<missing>}'" >&2
      failed=1
    fi
  done < <(module_files)

  if [[ "${found_modules}" != "true" ]]; then
    echo "no tracked go.mod or go.mod.src files found" >&2
    failed=1
  fi

  actual="$(awk '$1 == "go" { print $2; exit }' go.work)"
  if [[ "${actual}" != "${go_directive}" ]]; then
    echo "go.work: expected 'go ${go_directive}', found 'go ${actual:-<missing>}'" >&2
    failed=1
  fi

  actual="$(awk '$1 == "toolchain" { print $2; exit }' go.work)"
  if [[ "${actual}" != "${toolchain}" ]]; then
    echo "go.work: expected 'toolchain ${toolchain}', found 'toolchain ${actual:-<missing>}'" >&2
    failed=1
  fi

  while IFS= read -r image_ref; do
    [[ -n "${image_ref}" ]] || continue
    found_images=true
    if [[ "${image_ref}" != "golang:${patch}" ]]; then
      echo "expected 'golang:${patch}', found '${image_ref}'" >&2
      failed=1
    fi
  done < <(
    git --no-pager grep -hoE 'golang:1\.[0-9]+\.[0-9]+' -- \
      '*Dockerfile' \
      '*.Dockerfile' \
      '*.go' \
      '*.md' \
      '*.yaml' \
      '*.yml' || true
  )

  if [[ "${found_images}" != "true" ]]; then
    echo "no tracked golang image references found" >&2
    failed=1
  fi

  if [[ "${failed}" -ne 0 ]]; then
    return 1
  fi

  echo "All managed Go version references match ${patch}."
}

if [[ "${check_only}" == "true" ]]; then
  check_versions
  exit
fi

# 1. Keep all go.mod files and go.mod templates at the downstream-visible minimum version.
while IFS= read -r -d '' modfile; do
  go_edit mod edit -go="${go_directive}" "${modfile}"
  go_edit mod edit -fmt "${modfile}"
done < <(module_files)

# 2. Keep go.work at the minimum version, plus the actual toolchain version.
go_edit work edit -go="${go_directive}" -toolchain="${toolchain}" go.work
go_edit work edit -fmt go.work

# 3. Update golang:1.x.y references in Dockerfiles, docs, workflows, and generated source.
while IFS= read -r -d '' file; do
  replace_golang_image_versions "${file}"
done < <(image_files)

check_versions

echo
echo "Go version references after update:"
git --no-pager grep -nE \
  '(^go 1\.[0-9]+(\.[0-9]+)?$|toolchain go1\.[0-9]+\.[0-9]+|golang:1\.[0-9]+\.[0-9]+|go-version:)' \
  -- ':!:vendor/**' ':!:hack/update-go-version.sh' || true
