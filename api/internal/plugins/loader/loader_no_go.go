// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build no_go_plugins
// +build no_go_plugins

package loader

import (
	"fmt"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func (l *Loader) loadGoPlugin(id resid.ResId, absPath string) (resmap.Configurable, error) {
	return nil, fmt.Errorf("kustomize was built without go plugin support (-tags=no_go_plugins)")
}
