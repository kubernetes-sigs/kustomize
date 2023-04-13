// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import "sigs.k8s.io/kustomize/functions/examples/fn-framework-application/pkg/dispatcher"

func main() {
	_ = dispatcher.NewCommand().Execute()
}
