# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

for f in $(find $1 -name '*.go'); do
  echo $f
  # go run go.coder.com/go-tools/cmd/goimports
  ~/gopath/bin/goimports -w $f
done
