#!/bin/bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN=$(go env GOBIN)
MYGOBIN="${MYGOBIN:-$(go env GOPATH)/bin}"
VERSION=$1

cp $HOME/.kube/config /tmp/kubeconfig.txt | true
$MYGOBIN/kind create cluster --image kindest/node:$VERSION --name=getopenapidata

# TODO (natasha41575) Add a `kustomize openapi fetch --proto` option
kubectl proxy &
curl -k -H "Accept: application/com.github.proto-openapi.spec.v2@v1.0+protobuf" http://localhost:8001/openapi/v2 > /tmp/new_swagger.pb

$MYGOBIN/kind delete cluster --name=getopenapidata
cp /tmp/kubeconfig.txt $HOME/.kube/config | true
mkdir -p kubernetesapi/"${VERSION//.}"
cp /tmp/new_swagger.pb kubernetesapi/"${VERSION//.}"/swagger.pb
