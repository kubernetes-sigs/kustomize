// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

// This package imports tools required by build scripts.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// To update the pinned versions, update `go.mod` or call `go get URL@VERSION“
package tools

import (
	_ "sigs.k8s.io/kustomize/kustomize/v5"
)
