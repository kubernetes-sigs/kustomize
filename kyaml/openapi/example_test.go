// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi_test

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Example() {
	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "replicas")
	fmt.Println(f.Schema.Description[:70] + "...")
	fmt.Println(f.Schema.Type)

	// Output:
	// Number of desired pods. This is a pointer to distinguish between expli...
	// [integer]
}

func Example_arrayMerge() {
	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "template", "spec", "containers")
	fmt.Println(f.Schema.Description[:70] + "...")
	fmt.Println(f.Schema.Type)
	fmt.Println(f.PatchStrategyAndKey()) // merge patch strategy on name

	// Output:
	// List of containers belonging to the pod. Containers cannot currently b...
	// [array]
	// merge name
}

func Example_arrayReplace() {
	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "template", "spec", "containers", openapi.Elements, "args")
	fmt.Println(f.Schema.Description[:70] + "...")
	fmt.Println(f.Schema.Type)
	fmt.Println(f.PatchStrategyAndKey()) // no patch strategy or merge key

	// Output:
	// Arguments to the entrypoint. The docker image's CMD is used if this is...
	// [array]
}

func Example_arrayElement() {
	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("spec", "template", "spec", "containers",
		openapi.Elements, "ports", openapi.Elements, "containerPort")
	fmt.Println(f.Schema.Description[:70] + "...")
	fmt.Println(f.Schema.Type)

	// Output:
	// Number of port to expose on the pod's IP address. This must be a valid...
	// [integer]
}

func Example_map() {
	s := openapi.SchemaForResourceType(yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	f := s.Lookup("metadata", "labels")
	fmt.Println(f.Schema.Description[:70] + "...")
	fmt.Println(f.Schema.Type)

	// Output:
	// Map of string keys and values that can be used to organize and categor...
	// [object]
}
