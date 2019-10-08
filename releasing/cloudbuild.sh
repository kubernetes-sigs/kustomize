#!/bin/bash
set -e
set -x

# Script to run http://goreleaser.com

module=$1
shift

executable=$module
if [ "$module" == "api" ]; then
  # For this module, there's no correspondingly named
  # sub-directory, since the module is at the repo root.
  # There is, however, a dummy executable in a sub-directory.
  # Build that executable, primarily to give goreleaser
  # something to coordinate it release process around.
  executable=kustapiversion
fi

cd $executable

configFile=$(mktemp)
cat <<EOF >$configFile
project_name: $executable
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
- binary: $executable
  ldflags: >
    -s
    -X sigs.k8s.io/kustomize/v3/provenance.version={{.Version}}
    -X sigs.k8s.io/kustomize/v3/provenance.gitCommit={{.Commit}}
    -X sigs.k8s.io/kustomize/v3/provenance.buildDate={{.Date}}

  goos:
  - linux
  - darwin
  - windows
  goarch:
   - amd64
EOF

cat $configFile

/bin/goreleaser release --config=$configFile --rm-dist --skip-validate $@


