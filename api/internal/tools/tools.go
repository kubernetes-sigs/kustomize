// +build tools

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// This file exists to automatically trigger installs
// of the given tools, and is the official 'unofficial'
// way to declare a dependence on a Go binary until
// some better technique comes along.

package tools

import (
	// for code generation
	_ "golang.org/x/tools/cmd/stringer"
	// for lint checks
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// REMOVED pluginator from this process, and leaving
	// this note to discourage its reintroduction,
	// because pluginator depends on the api, forcing
	// major version increments in pluginator with each
	// api release to allow this trick to work and not
	// introduce cycles.
	// _ "sigs.k8s.io/kustomize/pluginator/v2"
)
