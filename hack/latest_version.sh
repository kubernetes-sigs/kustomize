#!/usr/bin/env bash
# Copyright 2024 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Looks up and prints the latest release version of the specified component.
# The first arg should be the component name (tag prefix).
#
# Requires GNU sort (or gsort) for version sorting.

set -euo pipefail

component="${1:-}"
if [ -z "${component}" ]; then
    echo >&2 "The first argument should be the requested component name."
    exit 2
fi

# Prefer gsort, if installed, to support coreutils install on mac with homebrew.
if command -v gsort >/dev/null 2>&1; then
    SORT="gsort"
elif command -v sort >/dev/null 2>&1; then
    SORT="sort"
else
    echo >&2 "sort (or gsort) required on the PATH for version sorting."
    exit 1
fi
# Validate sort has the -V or --version-sort option
if ${SORT} --help 2>&1 | grep -qe '\-\-version-sort'; then
    # GNU sort & BSD sort have both -V and --version-sort
    # Recent Mac sort has --version-sort, but not -V
    VERSION_FLAG="--version-sort"
elif ${SORT} --help 2>&1 | grep -qe '^[[:space:]]*\-V[[:space:]]*'; then
    # BusyBox sort has -V but not --version-sort
    VERSION_FLAG="-V"
else
    # Older Mac sort has neither -V nor --version-sort
    echo >&2 "sort (or gsort) on the PATH must support version sort (-V or --version-sort)."
    if "$(uname)" == "Darwin"; then
        echo >&2 "On Mac, either update the OS or use Homebrew to install the coreutils package, which includes GNU sort as gsort."
    elif "$(uname)" == "Linux"; then
        echo >&2 "On Linux, install the latest version of GNU or BSD sort."
    fi
    exit 1
fi

# We can't use /releases/latest because the desired component may not be the
# last component released.
RELEASES_URL="https://api.github.com/repos/kubernetes-sigs/kustomize/releases"

# You can authenticate by exporting the GITHUB_TOKEN in the environment
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    RELEASES_JSON=$(curl -s "$RELEASES_URL")
else
    RELEASES_JSON=$(curl -s "$RELEASES_URL" --header "Authorization: Bearer ${GITHUB_TOKEN}")
fi

if [[ "${RELEASES_JSON}" == *"API rate limit exceeded"* ]]; then
  echo "Github rate-limiter failed the request. Either authenticate or wait a couple of minutes."
  exit 1
fi

# This would be better with jq, but we're using grep and cut to avoid the
# dependency, even though it might be more fragile.
echo "${RELEASES_JSON}" \
    | grep -o "\"tag_name\": \"${component}/.*\"" \
    | cut -d\" -f4 | cut -d/ -f2 \
    | ${SORT} ${VERSION_FLAG} | tail -1
