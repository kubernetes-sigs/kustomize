// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// A dummy main to help with releasing the kustomize API module.
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/provenance"
)

// TODO: delete this when we find a better way to generate release notes.
func main() {
	fmt.Println(`
This 'main' exists only to make goreleaser create release notes for the API.
See https://github.com/goreleaser/goreleaser/issues/981
and https://github.com/kubernetes-sigs/kustomize/tree/master/releasing`)
	fmt.Println(provenance.GetProvenance())
}
