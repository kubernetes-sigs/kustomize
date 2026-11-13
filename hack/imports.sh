# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

for f in $(find $1 -name '*.go'); do
  echo $f
  go tool goimports -w $f
done
