// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi_test

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// The strategic-merge patch metadata for builtin Kubernetes types is served
// from a small precomputed table; no full schema document is embedded.
func Example() {
	s := openapi.PatchMetaSchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "template", "spec", "containers")
	fmt.Println(f.Schema.Type)
	fmt.Println(f.PatchStrategyAndKey()) // merge patch strategy on name

	// Output:
	// [array]
	// merge name
}

// Multi-key merges are expressed through the merge key list.
func Example_arrayMergeKeyList() {
	s := openapi.PatchMetaSchemaForResourceType(yaml.TypeMeta{APIVersion: "v1", Kind: "Service"})

	f := s.Lookup("spec", "ports")
	strategy, keys := f.PatchStrategyAndKeyList()
	fmt.Println(strategy)
	fmt.Println(keys)

	// Output:
	// merge
	// [port protocol]
}

// Fields that carry no strategic-merge metadata anywhere below them are not
// in the table; the merge walker treats a missing schema and a schema
// without patch extensions identically.
func Example_arrayReplace() {
	s := openapi.PatchMetaSchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "template", "spec", "containers", openapi.Elements, "args")
	fmt.Println(f == nil)

	// Output:
	// true
}

// Primitive lists can carry a merge strategy without a merge key.
func Example_primitiveMerge() {
	s := openapi.PatchMetaSchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("metadata", "finalizers")
	strategy, key := f.PatchStrategyAndKey()
	fmt.Printf("%q %q\n", strategy, key)

	// Output:
	// "merge" ""
}

// Supplying a schema document (the openapi field in kustomization.yaml, or
// AddSchema) restores full-fidelity lookups, descriptions included.
func Example_suppliedSchema() {
	err := openapi.AddSchema([]byte(`{
  "definitions": {
    "io.k8s.api.apps.v1.Deployment": {
      "x-kubernetes-group-version-kind": [{"group": "apps", "kind": "Deployment", "version": "v1"}],
      "properties": {
        "spec": {"description": "The desired behavior of the Deployment.", "type": "object"}
      }
    }
  }
}`))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer openapi.ResetOpenAPI()

	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	f := s.Lookup("spec")
	fmt.Println(f.Schema.Description)

	// Output:
	// The desired behavior of the Deployment.
}
