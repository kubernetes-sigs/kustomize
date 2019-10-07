#!/bin/bash
set -e
set -x

# Script to run http://goreleaser.com

module=$1
shift

if [ "$module" == "api" ]; then
  echo "goreleaser only releases 'main' packages (executables)"
  echo "See https://github.com/goreleaser/goreleaser/issues/981"
  exit 1
fi

config=$(mktemp)

cat <<EOF >$config
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
EOF

case "$module" in
  kustomize)
    cat <<EOF >>$config
builds:
- main: ./main.go
  binary: kustomize
  ldflags: -s -X sigs.k8s.io/kustomize/kustomize/v3/provenance.version={{.Version}} -X sigs.k8s.io/kustomize/kustomize/v3/provenance.gitCommit={{.Commit}} -X sigs.k8s.io/kustomize/kustomize/v3/provenance.buildDate={{.Date}}
  goos:
  - linux
  - darwin
  - windows
  goarch:
   - amd64
archive:
  format: binary
EOF
    ;;
  pluginator)
    cat <<EOF >>$config
builds:
- main: ./main.go
  binary: pluginator
  goos:
  - linux
  - darwin
  - windows
  goarch:
   - amd64
archive:
  format: binary
EOF
    ;;
  *)
    echo "Don't recognize module $module"
    exit 1
    ;;  
esac

cat $config

if [ "$module" != "api" ]; then
  # goreleaser must be run from the _module_ being released.
  cd $module
fi

/bin/goreleaser release --config=$config --rm-dist --skip-validate $@


