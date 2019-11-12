// +build tools

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// This file exists to trigger installs of the given tools.

package tools

import (
	// for code generation
	_ "golang.org/x/tools/cmd/stringer"
	// for lint checks
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// for integration tests driven by the examples
	_ "github.com/monopole/mdrip"
	// for generating code for builtin plugins
	_ "sigs.k8s.io/kustomize/pluginator/v2"
)
