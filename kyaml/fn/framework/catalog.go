// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

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
	KrmFunctions []KrmFunctionDefinitionSpec `yaml:"krmFunctions" json:"krmFunctions"`
}

// FindMatchingFunctionSpec accepts a catalog, and returns a KRMFunctionSpec that matches the
// metadata of the yaml node being passed.
func FindMatchingFunctionSpec(res *yaml.RNode, catalogs []Catalog) (*runtimeutil.FunctionSpec, error) {
	meta, err := res.GetMeta()
	if err != nil {
		return nil, fmt.Errorf("getting resource metadata: %w", err)
	}

	gv, err := schema.ParseGroupVersion(meta.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing group version: %w", err)
	}

	if res != nil {
		resGroup := gv.Group
		resKind := res.GetKind()
		resVersion := gv.Version

		for _, catalog := range catalogs {
			for _, krmFunc := range catalog.Spec.KrmFunctions {
				if krmFunc.Group == resGroup && krmFunc.Names.Kind == resKind {
					for _, version := range krmFunc.Versions {
						if version.Name == resVersion {
							return &version.Runtime, nil
						}
					}
				}
			}
		}
	}
	return nil, nil
}
