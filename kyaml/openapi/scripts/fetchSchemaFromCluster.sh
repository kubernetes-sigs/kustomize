#!/bin/bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN=$(go env GOPATH)/bin
VERSION=$1

cp $HOME/.kube/config /tmp/kubeconfig.txt | true
$MYGOBIN/kind create cluster --image kindest/node:$VERSION --name=getopenapidata
$MYGOBIN/kpt live  fetch-k8s-schema  --pretty-print > /tmp/new_swagger.json
$MYGOBIN/kind delete cluster --name=getopenapidata
cp /tmp/kubeconfig.txt $HOME/.kube/config | true
mkdir -p kubernetesapi/"${VERSION//.}"
cp /tmp/new_swagger.json kubernetesapi/"${VERSION//.}"/swagger.json
