#!/bin/bash
set -e
set -x

# goreleaser must be run from the _module_ being released.
cd kustomize

/bin/goreleaser \
  release \
  --config=../releasing/goreleaser.yaml \
  --rm-dist \
  --skip-validate \
  $@


