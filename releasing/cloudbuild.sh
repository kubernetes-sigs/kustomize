#!/bin/bash
set -e
set -x

# Script to run http://goreleaser.com

# Removed from `build` stanza
# binary: $module

module=$1
shift

# The following assumes git tags formatted like
# "api/v1.2.3" and splits on the slash.
# Goreleaser doesn't know what to do with this
# tag format, and fails when creating an archive
# with a / in the name.
fullTag=$(git describe)
export tModule=${fullTag%/*}
export tSemver=${fullTag#*/}
echo "tModule=$tModule"
echo "tSemver=$tSemver"
if [ "$module" != "$tModule" ]; then
  # Tag and argument sanity check
  echo "Unexpected mismatch: moduleFromArg=$module, moduleFromTag=$tModule"
  echo "Either the module arg to this script is wrong, or the git tag is wrong."
  exit 1
fi

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
  draft: true
builds:
- ldflags: >
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
archives:
-  name_template: "${module}_${tSemver}_{{ .Os }}_{{ .Arch }}"
EOF

cat $configFile

/bin/goreleaser release --config=$configFile --rm-dist --skip-validate $@
