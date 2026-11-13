# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Usage:  From repo root:
#  ./hack/doGoMod.sh tidy
#  ./hack/doGoMod.sh verify

operation=$1
# site/themes/docsy is a third-party Hugo theme submodule and is not part of
# go.work. Its go.mod requires a Font-Awesome version the module proxy no
# longer serves, which makes `make workspace-sync` fail before tests run.
for f in $(find ./ -name 'go.mod' -not -path './site/themes/*'); do
  echo $f
  d=$(dirname "$f")
  (cd $d; go mod $operation)
done
