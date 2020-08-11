// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

// ResourceIdentifier contains the information needed to uniquely
// identify a resource in a cluster.
type ResourceIdentifier struct {
	// Name is the name of the resource as set in metadata.name
	Name string `yaml:"name,omitempty"`
	// Namespace is the namespace of the resource as set in metadata.namespace
	Namespace string `yaml:"namespace,omitempty"`
	// ApiVersion is the apiVersion of the resource
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind is the kind of the resource
	Kind string `yaml:"kind,omitempty"`
}

func (r *ResourceIdentifier) GetName() string {
	return r.Name
}

func (r *ResourceIdentifier) GetNamespace() string {
	return r.Namespace
}

func (r *ResourceIdentifier) GetAPIVersion() string {
	return r.APIVersion
}

func (r *ResourceIdentifier) GetKind() string {
	return r.Kind
}
