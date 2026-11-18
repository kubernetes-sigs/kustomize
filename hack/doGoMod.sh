# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Usage:  From repo root:
#  ./hack/doGoMod.sh tidy
#  ./hack/doGoMod.sh verify

operation=$1
# The site is an external submodule, not part of the Kustomize Go workspace.
for f in $(find ./ -name 'go.mod' -not -path './site/*'); do
  echo $f
  d=$(dirname "$f")
  (cd $d; go mod $operation)
done
