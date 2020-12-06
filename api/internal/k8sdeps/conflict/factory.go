// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package conflict

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	sp "k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resource"
)

type cdFactory struct {
	rf *resource.Factory
}

var _ resource.ConflictDetectorFactory = &cdFactory{}

// NewFactory returns a conflict detector factory.
// The detector uses a resource factory to convert resources to/from
// json/yaml/maps representations.
func NewFactory(rf *resource.Factory) resource.ConflictDetectorFactory {
	return &cdFactory{rf: rf}
}

// New returns a conflict detector that's aware of the GVK type.
func (f *cdFactory) New(gvk resid.Gvk) (resource.ConflictDetector, error) {
	// Convert to apimachinery representation of object
	obj, err := scheme.Scheme.New(schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	})
	if err == nil {
		meta, err := sp.NewPatchMetaFromStruct(obj)
		return &conflictDetectorSm{
			lookupPatchMeta: meta, resourceFactory: f.rf}, err
	}
	if runtime.IsNotRegisteredError(err) {
		return &conflictDetectorJson{resourceFactory: f.rf}, nil
	}
	return nil, err
}
