// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapandsecret

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
)

// Factory makes ConfigMaps and Secrets.
type Factory struct {
	kvLdr ifc.KvLoader
}

// NewFactory returns a new factory that makes ConfigMaps and Secrets.
func NewFactory(kvLdr ifc.KvLoader) *Factory {
	return &Factory{kvLdr: kvLdr}
}

// copyLabelsAndAnnotations copies labels and annotations from
// GeneratorOptions into the given object.
func (f *Factory) copyLabelsAndAnnotations(
	obj metav1.Object, opts *types.GeneratorOptions) {
	if opts == nil {
		return
	}
	if opts.Labels != nil {
		obj.SetLabels(types.CopyMap(opts.Labels))
	}
	if opts.Annotations != nil {
		obj.SetAnnotations(types.CopyMap(opts.Annotations))
	}
}

const keyExistsErrorMsg = "cannot add key %s, another key by that name already exists: %v"
