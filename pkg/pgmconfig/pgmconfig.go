// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package pgmconfig holds global constants for the kustomize tool.
package pgmconfig

// RecognizedKustomizationFileNames is a list of file names
// that kustomize recognizes.
// To avoid ambiguity, a kustomization directory may not
// contain more than one match to this list.
func RecognizedKustomizationFileNames() []string {
	return []string{
		"kustomization.yaml",
		"kustomization.yml",
		"Kustomization",
	}
}

func DefaultKustomizationFileName() string {
	return RecognizedKustomizationFileNames()[0]
}

const (
	// An environment variable to consult for kustomization
	// configuration data.  See:
	// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
	XdgConfigHome = "XDG_CONFIG_HOME"

	// Use this when XdgConfigHome not defined.
	DefaultConfigSubdir = ".config"

	// Program name, for help, finding the XDG_CONFIG_DIR, etc.
	ProgramName = "kustomize"

	// Domain from which kustomize code is imported, for locating
	// plugin source code under $GOPATH.
	// TODO: move to pgk/plugin/config.go or equivalent
	// as part of v4 release.  Cannot move till then
	// because of pluginator dependence at v3.
	DomainName = "sigs.k8s.io"

	// Name of directory housing all plugins.
	// TODO: move to pgk/plugin/config.go or equivalent
	PluginRoot = "plugin"
)
