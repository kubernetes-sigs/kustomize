// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main contains a function to be used for e2e testing.
//
// The function is written using the framework, and parses the ResourceList.functionConfig
// into a go struct.
//
// The function will set 3 annotations on each resource.
//
// See https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
package main
