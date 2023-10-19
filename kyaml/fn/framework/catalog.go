// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import "sigs.k8s.io/kustomize/kyaml/yaml"

const CatalogKind = "Catalog"

// Catalog holds a collection of one or more functions
// that can be used with a Kustomize resource.
type Catalog struct {
	// APIVersion and Kind of the object.
	// Must be config.kubernetes.io/v1alpha1 and Catalog respectively.
	yaml.TypeMeta `json:",inline" yaml:",inline"`

	// Standard KRM object metadata.
	yaml.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Spec contains a list of KRMFunctionDefinetions.
	Spec CatalogSpec `yaml:"spec" json:"spec"`
}

type CatalogSpec struct {
	KrmFunctions []KRMFunctionDefinition `yaml:"krmFunctions" json:"krmFunctions"`
}
