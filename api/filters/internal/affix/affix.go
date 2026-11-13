// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package affix

import (
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Apply adds a prefix or suffix to scalar fields selected by the field spec.
func Apply(nodes []*yaml.RNode, fieldSpec types.FieldSpec, prefix, suffix string, setter filtersutil.TrackableSetter) ([]*yaml.RNode, error) {
	filter := kio.FilterAll(yaml.FilterFunc(func(node *yaml.RNode) (*yaml.RNode, error) {
		err := node.PipeE(fieldspec.Filter{
			FieldSpec: fieldSpec,
			SetValue: func(field *yaml.RNode) error {
				return setter.SetScalar(prefix + field.YNode().Value + suffix)(field)
			},
			CreateKind: yaml.ScalarNode,
			CreateTag:  yaml.NodeTagString,
		})
		return node, err //nolint:wrapcheck // Preserve the error chain already wrapped by PipeE.
	}))
	return filter.Filter(nodes) //nolint:wrapcheck // Preserve the prefix and suffix filters' existing error chain.
}
