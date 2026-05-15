#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

patch="${1:?usage: $0 <go patch version: 1.N.M>}"

if [[ ! "${patch}" =~ ^1\.[0-9]+\.[0-9]+$ ]]; then
  echo "patch must look like 1.N.M, e.g. 1.25.10" >&2
  exit 2
fi

minor="${patch%.*}"
go_directive="${minor}.0"
toolchain="go${patch}"

echo "go directive       : ${go_directive}"
echo "toolchain / Docker : ${patch}"

# 1. Keep all go.mod files and go.mod templates at the downstream-visible minimum version.
find . \
  \( -name go.mod -o -name go.mod.src \) \
  -not -path './.git/*' \
  -not -path './vendor/*' \
  -print0 |
while IFS= read -r -d '' modfile; do
  go mod edit -go="${go_directive}" "${modfile}"
  go mod edit -fmt "${modfile}"
done

# 2. Keep go.work at the minimum version, plus the actual toolchain version.
if [[ -f go.work ]]; then
  go work edit -go="${go_directive}" -toolchain="${toolchain}" go.work
  go work edit -fmt go.work
fi

# 3. Update golang:1.x.y references in Dockerfiles, docs, workflows, and generated source.
git --no-pager grep -zlE 'golang:1\.[0-9]+\.[0-9]+' -- \
  '*Dockerfile' \
  '*.Dockerfile' \
  '*.go' \
  '*.md' \
  '*.yaml' \
  '*.yml' |
while IFS= read -r -d '' file; do
  perl -0pi -e "s#(golang:)1\\.[0-9]+\\.[0-9]+#\${1}${patch}#g" "${file}"
done

echo
echo "Go version references after update:"
git --no-pager grep -nE \
  '(^go 1\.[0-9]+(\.[0-9]+)?$|toolchain go1\.[0-9]+\.[0-9]+|golang:1\.[0-9]+\.[0-9]+|go-version:)' \
  -- ':!:vendor/**' ':!:hack/update-go-version.sh' || true
