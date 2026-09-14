#!/usr/bin/env bash
# Copyright 2026 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source "$(dirname "${BASH_SOURCE[0]}")/install_kustomize.sh"

failures=0

assert_tag() {
  local version=$1
  local expected=$2
  local actual
  actual=$(release_tag_for_version "$version")
  if [[ "$actual" != "$expected" ]]; then
    echo "release_tag_for_version $version: expected $expected, got $actual"
    failures=$((failures + 1))
  fi
}

assert_tag 2.0.3 v2.0.3
assert_tag 3.0.0 v3.0.0
assert_tag 3.2.0 v3.2.0
assert_tag 3.2.1 kustomize/v3.2.1
assert_tag 3.2.2 kustomize/v3.2.2
assert_tag 3.3.0 kustomize/v3.3.0
assert_tag 5.8.1 kustomize/v5.8.1
assert_tag 10.0.0 kustomize/v10.0.0

if ((failures > 0)); then
  echo "$failures assertion(s) failed"
  exit 1
fi

echo "all release_tag_for_version assertions passed"
