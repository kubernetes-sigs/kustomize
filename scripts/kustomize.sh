#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# kustomize.sh: a wrapper for the kubernetes binary used to make it easier to deploy
#
# the script will:
# - download the appropriate binary from github
# - verify that the checksum is ok
# - cache the binary for further executions

set -euo pipefail

RELEASE=1.0.7

# See here for the checksums below:
# https://github.com/kubernetes-sigs/kustomize/releases/download/v1.0.6/checksums.txt
CHECKSUM_LINUX=afac1af5a6a7c688d5665f729d4d3a48492836f5baf20ceb1e7b71211a626ca5
CHECKSUM_DARWIN=28d0265d4139f65dbd2c01fb91da63fc13e7bba4fd3e19f41595a02a453da2c4
CHECKSUM_WINDOWS=2e9b1b0600c906760a0fe04136185fb44356fab32605a7524978c0e6345fea1b
CACHE_DIR=$HOME/.cache/kustomize
RELEASE_DOWNLOAD_BASE=https://github.com/kubernetes-sigs/kustomize/releases/download/v$RELEASE

# Detect my architecture
UNAME=$(uname -ms)
ARCH=""
case "$UNAME" in
	"Darwin x86_64")
		ARCH="darwin_amd64"
		CHECKSUM_EXPECTED="$CHECKSUM_DARWIN"
		sha256() {
			shasum -a 256 $1 | awk '{print $1}'
		}
		;;
	"Linux x86_64")
		ARCH="linux_amd64"
		CHECKSUM_EXPECTED="$CHECKSUM_LINUX"
		sha256() {
			sha256sum $1 | awk '{print $1}'
		}
		;;
	MINGW*)
		ARCH="windows_amd64"
		CHECKSUM_EXPECTED="$CHECKSUM_WINDOWS"
		sha256() {
			sha256sum $1 | awk '{print $1}'
		}
		;;
	*)
		echo "ERROR: unknown architecture: $UNAME" >&2
		exit 1
		;;
esac

verify_checksum() {
	printf "verifying checksum..." >&2
	sha256_new=$(sha256 $1)
	if [[ $sha256_new != $2 ]]; then
		echo "ERROR: wrong checksum of downloaded binary $1 (expected: $2, is: $sha256_new)" >&2
		exit 1
	fi
	echo " ok" >&2
}


# Determine cached binary name
BINARY_NAME="kustomize_${RELEASE}_$ARCH"
if [[ $ARCH = "windows_amd64" ]]; then BINARY_NAME="$BINARY_NAME.exe"; fi
CACHED_BINARY="$CACHE_DIR/$BINARY_NAME"

# Download if needed
if [[ ! -e "$CACHED_BINARY" ]]; then
	mkdir -p $CACHE_DIR
	echo "downloading $BINARY_NAME..." >&2
	curl --fail --location --progress-bar $RELEASE_DOWNLOAD_BASE/$BINARY_NAME >$CACHED_BINARY.$$
	# Verify checksum
	verify_checksum "$CACHED_BINARY.$$" "$CHECKSUM_EXPECTED"
	chmod +x $CACHED_BINARY.$$
	mv $CACHED_BINARY.$$ $CACHED_BINARY
fi

# Execute
exec $CACHED_BINARY "$@"
