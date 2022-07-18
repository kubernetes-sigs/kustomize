#!/bin/bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN=$(go env GOBIN)
MYGOBIN="${MYGOBIN:-$(go env GOPATH)/bin}"
VERSION=$1

$MYGOBIN/go-bindata \
  --pkg "${VERSION//./_}" \
  -o kubernetesapi/"${VERSION//./_}"/swagger.go \
  kubernetesapi/"${VERSION//./_}"/swagger.pb
