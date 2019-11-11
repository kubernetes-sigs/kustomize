// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// A dummy main to help with releasing the pseudo/k8s module.
package main

import (
	"fmt"
)

// TODO: delete this when we find a better way to generate release notes.
func main() {
	fmt.Println(`
This 'main' exists to help goreleaser create release notes for the pseudo/k8s module.
See https://github.com/goreleaser/goreleaser/issues/981
and https://github.com/kubernetes-sigs/kustomize/tree/master/releasing`)
}
