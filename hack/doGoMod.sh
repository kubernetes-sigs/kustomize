# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Usage:  From repo root:
#  ./hack/doGoMod.sh tidy
#  ./hack/doGoMod.sh verify

operation=$1
for f in $(find ./ -name 'go.mod'); do
  echo $f
  d=$(dirname "$f")
  (cd $d; go mod $operation)
done
