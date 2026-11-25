#!/usr/bin/env bash
# Copyright 2026 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Pulls the base images used by the docker e2e test Dockerfiles before the
# tests run, so a registry rate limit (e.g. a 429 "Data limit exceeded" from
# public.ecr.aws) is retried here instead of failing the tests.
#
# `docker build` without --pull uses a base image that is already present
# locally, so once an image is pulled (or retagged from the Docker Hub
# fallback) the tests no longer contact the registry for it.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
attempts="${PREPULL_ATTEMPTS:-5}"

pull_with_retry() {
  local image="$1"
  local i
  for ((i = 1; i <= attempts; i++)); do
    if docker pull --quiet "$image"; then
      return 0
    fi
    if ((i < attempts)); then
      echo "pull of $image failed (attempt $i/$attempts), retrying in $((i * 15))s" >&2
      sleep $((i * 15))
    fi
  done
  return 1
}

images=$(grep -h '^FROM' "$repo_root"/cmd/config/internal/commands/e2e/*/Dockerfile |
  awk '{print $2}' | sort -u)

for image in $images; do
  echo "pre-pulling $image"
  if pull_with_retry "$image"; then
    continue
  fi
  # public.ecr.aws/docker/library mirrors the Docker Hub official images.
  if [[ "$image" == public.ecr.aws/docker/library/* ]]; then
    fallback="docker.io/library/${image#public.ecr.aws/docker/library/}"
    echo "falling back to $fallback" >&2
    pull_with_retry "$fallback"
    docker tag "$fallback" "$image"
    continue
  fi
  echo "failed to pull $image" >&2
  exit 1
done
