#!/bin/bash
# Copyright 2024 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Build the release binaries for every OS/arch combination.
# It builds compressed artifacts on $release_dir.
function build_kustomize_binary {
  echo "build kustomize binaries"
  
  version=$1
  release_dir=$2

  echo "build release artifacts to $release_dir"

  mkdir -p "output"
  
  # build date in ISO8601 format
  build_date=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  for os in linux darwin windows; do
    arch_list=(amd64 arm64)
    if [ "$os" == "linux" ]; then
      arch_list=(amd64 arm64 s390x ppc64le)
    fi
    for arch in "${arch_list[@]}" ; do
      echo "Building $os-$arch"
    #   CGO_ENABLED=0 GOWORK=off GOOS=$os GOARCH=$arch go build -o output/kustomize -ldflags\
      binary_name="kustomize"
      [[ "$os" == "windows" ]] && binary_name="kustomize.exe"
      CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o output/$binary_name -ldflags\
        "-s -w\
        -X sigs.k8s.io/kustomize/api/provenance.version=$version\
        -X sigs.k8s.io/kustomize/api/provenance.gitCommit=$(git rev-parse HEAD)\
        -X sigs.k8s.io/kustomize/api/provenance.buildDate=$build_date"\
        kustomize/main.go
      if [ "$os" == "windows" ]; then
        zip -j "${release_dir}/kustomize_${version}_${os}_${arch}.zip" output/$binary_name
      else
        tar cvfz "${release_dir}/kustomize_${version}_${os}_${arch}.tar.gz" -C output $binary_name
      fi
      rm output/$binary_name
    done
  done

  # create checksums.txt
  pushd "${release_dir}"
  for release in *; do
    echo "generate checksum: $release"
    sha256sum "$release" >> checksums.txt
  done
  popd

  rmdir output
}

main() {

  currentTag=$(git describe --tags)
  version=${currentTag##*/}
  
  if grep -q -E '^[0-9]+(\.[0-9]+)*$' <<< "$version"
  then
      printf "%s is NOT in semver format.\n" "$version"
      exit 1
  fi

  mkdir -p dist
  release_artifact_dir=${PWD}/dist
  build_kustomize_binary "${version}" "${release_artifact_dir}"
}

main $@