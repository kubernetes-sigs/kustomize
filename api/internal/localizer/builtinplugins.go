// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import "sigs.k8s.io/kustomize/kyaml/yaml"

// localizeBuiltinPlugin localizes built-in plugins with file paths.
type localizeBuiltinPlugin struct {
}

// Filter localizes the built-in plugins with file paths. Filter returns an error if
// plugins contains a resource that is not a built-in plugin, cannot contain a file path,
// or is not localizable.
// TODO(annasong): implement
func (lbp *localizeBuiltinPlugin) Filter(plugins []*yaml.RNode) ([]*yaml.RNode, error) {
	return plugins, nil
}
