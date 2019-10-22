#!/bin/bash
set -e
set -x

# Script to run http://goreleaser.com

module=$1
shift

cd $module

configFile=$(mktemp)
cat <<EOF >$configFile
project_name: $module
env:
- CGO_ENABLED=0
- GO111MODULE=on
checksum:
  name_template: 'checksums.txt'
changelog:
  sort: asc
  filters:
    exclude:
    - '^docs:'
    - '^test:'
    - Merge pull request
    - Merge branch
release:
  github:
    owner: kubernetes-sigs
    name: kustomize
builds:
- binary: $module
  ldflags: >
    -s
    -X sigs.k8s.io/kustomize/api/provenance.version={{.Version}}
    -X sigs.k8s.io/kustomize/api/provenance.gitCommit={{.Commit}}
    -X sigs.k8s.io/kustomize/api/provenance.buildDate={{.Date}}

  goos:
  - linux
  - darwin
  - windows
  goarch:
   - amd64
EOF

cat $configFile

/bin/goreleaser release --config=$configFile --rm-dist --skip-validate $@
