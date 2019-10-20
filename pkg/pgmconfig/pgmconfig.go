// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package pgmconfig holds global constants for the kustomize tool.
package pgmconfig

const (

	// Program name, for help, finding the XDG_CONFIG_DIR, etc.
	ProgramName = "kustomize"

	// TODO: delete this.  it's a copy of a const
	// defined elsewhere but used by pluginator.
	DomainName = "sigs.k8s.io"

	// TODO: delete this.  its a copy of a const
	// defined elsewhere but used by pluginator.
	PluginRoot = "plugin"
)
