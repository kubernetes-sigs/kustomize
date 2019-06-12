// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package transformers

import (
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

func TestPrefixSuffixNameRun(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	m := resmap.FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm2",
				},
			}),
		resid.NewResId(crd, "crd"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
	})
	expected := resmap.FromMap(map[resid.ResId]*resource.Resource{
		resid.NewResIdWithPrefixSuffix(cmap, "cm1", "someprefix-", "-somesuffix"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm1-somesuffix",
				},
			}),
		resid.NewResIdWithPrefixSuffix(cmap, "cm2", "someprefix-", "-somesuffix"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm2-somesuffix",
				},
			}),
		resid.NewResId(crd, "crd"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1beta1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crd",
				},
			}),
	})

	npst, err := NewPrefixSuffixTransformer(
		"someprefix-", "-somesuffix", defaultTransformerConfig.NamePrefix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = npst.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = expected.ErrorIfNotEqualSets(m); err != nil {
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
