// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// The build tag "kustomize_disable_go_plugin_support" is used to deactivate the
// kustomize API's dependency on the "plugins" package. This is beneficial for
// applications that need to embed it but do not have requirements for dynamic
// Go plugins.
// Including plugins as a dependency can lead to an increase in binary size due
// to the population of ELF's sections such as .dynsym and .dynstr.
// By utilizing this flag, applications have the flexibility to exclude the
// import if they do not require support for dynamic Go plugins.
//go:build kustomize_disable_go_plugin_support

package loader

import (
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func (l *Loader) loadGoPlugin(id resid.ResId, absPath string) (resmap.Configurable, error) {
	return nil, errors.New("plugin load is disabled")
}
