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
	fmt.Println(`This 'main' exists only to create release notes for the API.`)
	fmt.Println(provenance.GetProvenance())
}
