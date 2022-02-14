// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// TODO: This should eventually replace the openapi package when it is complete.
package openapignostic

import (
	"fmt"
	"path/filepath"
	"strings"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kustomizationapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// openApiSchema contains the parsed openapi state.  this is in a struct rather than
// a list of vars so that it can be reset from tests.
type openApiSchema struct {
	// schema holds the OpenAPI schema data
	schema openapi_v2.Document

	// schemaForResourceType is a map of Resource types to their schemas
	schemaByResourceType map[yaml.TypeMeta]*openapi_v2.NamedSchema

	// namespaceabilityByResourceType stores whether a given Resource type
	// is namespaceable or not
	namespaceabilityByResourceType map[yaml.TypeMeta]bool

	// noUseBuiltInSchema stores whether we want to prevent using the built-n
	// Kubernetes schema as part of the global schema
	noUseBuiltInSchema bool

	// schemaInit stores whether or not we've parsed the schema already,
	// so that we only reparse the schema when necessary (to speed up performance)
	schemaInit bool
}

var globalOpenApiSchema openApiSchema

// parseBuiltin calls parseDocument to parse the json schemas
func parseBuiltin(version string) {
	if globalOpenApiSchema.noUseBuiltInSchema {
		// don't parse the built in schema
		return
	}

	// parse the swagger, this should never fail
	assetName := filepath.Join(
		"kubernetesapi",
		version,
		"swagger.json")

	if err := parseDocument(kubernetesapi.OpenAPIMustAsset[version](assetName)); err != nil {
		// this should never happen
		panic(err)
	}

	if err := parseDocument(kustomizationapi.MustAsset("kustomizationapi/swagger.json")); err != nil {
		// this should never happen
		panic(err)
	}
}

// parseDocument parses and indexes a single json schema
func parseDocument(b []byte) error {
	doc, err := openapi_v2.ParseDocument(b)
	if err != nil {
		return errors.Wrap(fmt.Errorf("parsing document error: %w", err))
	}
	AddOpenApiDefinitions(doc.Definitions)
	findNamespaceabilityFromOpenApi(doc.Paths)

	return nil
}

// SchemaForResourceType returns the schema for a provided GVK
func SchemaForResourceType(meta yaml.TypeMeta) *openapi_v2.NamedSchema{
	return globalOpenApiSchema.schemaByResourceType[meta]
}

// AddDefinitions adds the definitions to the global schema.
func AddOpenApiDefinitions(definitions *openapi_v2.Definitions) {
	// initialize values if they have not yet been set
	if globalOpenApiSchema.schemaByResourceType == nil {
		globalOpenApiSchema.schemaByResourceType = map[yaml.TypeMeta]*openapi_v2.NamedSchema{}
	}
	if globalOpenApiSchema.schema.Definitions == nil {
		globalOpenApiSchema.schema.Definitions = &openapi_v2.Definitions{}
	}

	props := definitions.AdditionalProperties
	// index the schema definitions so we can lookup them up for Resources
	for k := range props {
		// index by GVK, if no GVK is found then it is the schema for a subfield
		// of a Resource
		d := props[k]

		// copy definitions to the schema
		globalOpenApiSchema.schema.Definitions.AdditionalProperties = append(globalOpenApiSchema.schema.Definitions.AdditionalProperties, d)

		for _, e := range d.GetValue().GetVendorExtension() {
			if e.Name == "x-kubernetes-group-version-kind" {
				var exts []map[string]string
				if err := yaml.Unmarshal([]byte(e.GetValue().GetYaml()), &exts); err != nil {
					continue
				}
				for _, gvk := range exts {
					typeMeta := yaml.TypeMeta{
						APIVersion: strings.Trim(strings.Join([]string{gvk["group"], gvk["version"]}, "/"), "/"),
						Kind:       gvk["kind"],
					}
					globalOpenApiSchema.schemaByResourceType[typeMeta] = d
				}
			}
		}
	}
}

// findNamespaceability looks at the api paths for the resource to determine
// if it is cluster-scoped or namespace-scoped. The gvk of the resource
// for each path is found by looking at the x-kubernetes-group-version-kind
// extension. If a path exists for the resource that contains a namespace path
// parameter, the resource is namespace-scoped.
func findNamespaceabilityFromOpenApi(paths *openapi_v2.Paths) {
	if globalOpenApiSchema.namespaceabilityByResourceType == nil {
		globalOpenApiSchema.namespaceabilityByResourceType = make(map[yaml.TypeMeta]bool)
	}
	if paths == nil {
		return
	}
	for _, p := range paths.Path {
		path, pathInfo := p.GetName(), p.GetValue()
		if pathInfo.GetGet() == nil {
			continue
		}
		for _, e := range pathInfo.GetGet().GetVendorExtension() {
			if e.Name == "x-kubernetes-group-version-kind" {
				var gvk map[string]string
				if err := yaml.Unmarshal([]byte(e.GetValue().GetYaml()), &gvk); err != nil {
					continue
				}
				typeMeta := yaml.TypeMeta{
					APIVersion: strings.Trim(strings.Join([]string{gvk["group"], gvk["version"]}, "/"), "/"),
					Kind:       gvk["kind"],
				}
				if strings.Contains(path, "namespaces/{namespace}") {
					// if we find a namespace path parameter, we just update the map
					// directly
					globalOpenApiSchema.namespaceabilityByResourceType[typeMeta] = true
				} else if _, found := globalOpenApiSchema.namespaceabilityByResourceType[typeMeta]; !found {
					// if the resource doesn't have the namespace path parameter, we
					// only add it to the map if it doesn't already exist.
					globalOpenApiSchema.namespaceabilityByResourceType[typeMeta] = false
				}
			}
		}
	}
}

// ResetOpenAPI resets the openapi data to empty
func ResetSchema() {
	globalOpenApiSchema = openApiSchema{}
}
