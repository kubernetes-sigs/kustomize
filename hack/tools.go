// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

// This package imports tools required by build scripts.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// To update the pinned versions, update `go.mod` or call `go get URL@VERSION“
package hack

import (
	// for code generation
	_ "golang.org/x/tools/cmd/stringer"
	// for code formatting
	_ "golang.org/x/tools/cmd/goimports"
	// for code linting
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// for site serving
	_ "github.com/gohugoio/hugo"
	// for testable markdown examples
	_ "github.com/monopole/mdrip"
	// for cobra command help text generation from markdown
	_ "sigs.k8s.io/kustomize/cmd/mdtogo"
	// for license file header injection
	_ "github.com/google/addlicense"
	// for local testing
	_ "sigs.k8s.io/kind"
	// for embeding code and manifest into markdown docs
	_ "github.com/campoy/embedmd"
	// for generating CRDs
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	// for bundling non-go files into go binaries
	_ "github.com/go-bindata/go-bindata/v3/go-bindata"
	// for checking Go API compatibility
	_ "github.com/joelanford/go-apidiff"
	// for creating GitHub PRs & releases
	_ "github.com/cli/cli/cmd/gh"
	// for validating k8s manifests
	_ "github.com/instrumenta/kubeval"
)
