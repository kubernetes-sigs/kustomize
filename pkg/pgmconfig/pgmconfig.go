// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package pgmconfig holds global constants for the kustomize tool.
package pgmconfig

// KustomizationFileNames is a list of filenames
// that kustomize recognizes.
// To avoid ambiguity, a directory cannot contain
// more than one match to this list.
func KustomizationFileNames() []string {
	return []string{
		KustomizationFileName0,
		KustomizationFileName1,
		KustomizationFileName2}
}

const (
	KustomizationFileName0 = "kustomization.yaml"
	KustomizationFileName1 = "kustomization.yml"
	KustomizationFileName2 = "Kustomization"

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
