// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	_ "embed"
	"maps"
	"slices"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

//go:generate go run ./cmd/openapi-bundle -input kubernetesapi/v1_21_2/swagger.pb.gz -output kubernetesapi/data/kubernetes-openapi-union-v1.21.2.bundle-v1.json -kubernetes-version v1.21.2

const (
	// DefaultOpenAPI is the Kubernetes version represented by the built-in
	// schema. It remains v1.21.2 during the initial artifact-format migration.
	DefaultOpenAPI = "v1.21.2"

	// BuiltinSchemaInfo is the value printed by `kustomize openapi info`.
	BuiltinSchemaInfo = "{title:Kubernetes,version:" + DefaultOpenAPI + "}"
)

//go:embed kubernetesapi/data/kubernetes-openapi-union-v1.21.2.bundle-v1.json
var builtinKubernetesOpenAPIBundle []byte

//go:embed kustomizationapi/swagger.json
var builtinKustomizationOpenAPI []byte

func hasBuiltinOpenAPIVersion(version string) bool {
	return version == DefaultOpenAPI
}

// Only decoded JSON values enter these copies; references form strings rather
// than object cycles, and extension values are maps, slices or scalars.
func copyBuiltinValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		result := maps.Clone(value)
		for key, item := range result {
			result[key] = copyBuiltinValue(item)
		}
		return result
	case []any:
		result := slices.Clone(value)
		for i, item := range result {
			result[i] = copyBuiltinValue(item)
		}
		return result
	default:
		return value
	}
}

func copyBuiltinSchemas[M ~map[string]spec.Schema](schemas M) M {
	result := maps.Clone(schemas)
	for key, schema := range result {
		result[key] = copyBuiltinSchema(schema)
	}
	return result
}

func copyBuiltinSchema(schema spec.Schema) spec.Schema {
	schema.Extensions = maps.Clone(schema.Extensions)
	for key, value := range schema.Extensions {
		schema.Extensions[key] = copyBuiltinValue(value)
	}
	schema.ExtraProps = maps.Clone(schema.ExtraProps)
	for key, value := range schema.ExtraProps {
		schema.ExtraProps[key] = copyBuiltinValue(value)
	}
	if ref := schema.Ref.String(); ref != "" || schema.Ref.IsRoot() {
		schema.Ref = spec.MustCreateRef(ref)
	}
	schema.Type = slices.Clone(schema.Type)
	schema.Required = slices.Clone(schema.Required)
	schema.Default = copyBuiltinValue(schema.Default)
	schema.Example = copyBuiltinValue(schema.Example)
	schema.Enum = slices.Clone(schema.Enum)
	for i, value := range schema.Enum {
		schema.Enum[i] = copyBuiltinValue(value)
	}
	schema.Maximum = copyBuiltinPointer(schema.Maximum)
	schema.Minimum = copyBuiltinPointer(schema.Minimum)
	schema.MaxLength = copyBuiltinPointer(schema.MaxLength)
	schema.MinLength = copyBuiltinPointer(schema.MinLength)
	schema.MaxItems = copyBuiltinPointer(schema.MaxItems)
	schema.MinItems = copyBuiltinPointer(schema.MinItems)
	schema.MultipleOf = copyBuiltinPointer(schema.MultipleOf)
	schema.MaxProperties = copyBuiltinPointer(schema.MaxProperties)
	schema.MinProperties = copyBuiltinPointer(schema.MinProperties)
	schema.ExternalDocs = copyBuiltinPointer(schema.ExternalDocs)
	schema.Properties = copyBuiltinSchemas(schema.Properties)
	schema.PatternProperties = copyBuiltinSchemas(schema.PatternProperties)
	schema.Definitions = copyBuiltinSchemas(schema.Definitions)
	schema.AllOf = copyBuiltinSchemaSlice(schema.AllOf)
	schema.OneOf = copyBuiltinSchemaSlice(schema.OneOf)
	schema.AnyOf = copyBuiltinSchemaSlice(schema.AnyOf)
	if schema.Not != nil {
		copy := copyBuiltinSchema(*schema.Not)
		schema.Not = &copy
	}
	if schema.Items != nil {
		schema.Items = copyBuiltinPointer(schema.Items)
		schema.Items.Schemas = copyBuiltinSchemaSlice(schema.Items.Schemas)
		if schema.Items.Schema != nil {
			copy := copyBuiltinSchema(*schema.Items.Schema)
			schema.Items.Schema = &copy
		}
	}
	schema.AdditionalProperties = copyBuiltinSchemaOrBool(schema.AdditionalProperties)
	schema.AdditionalItems = copyBuiltinSchemaOrBool(schema.AdditionalItems)
	schema.Dependencies = maps.Clone(schema.Dependencies)
	for key, dependency := range schema.Dependencies {
		dependency.Property = slices.Clone(dependency.Property)
		if dependency.Schema != nil {
			copy := copyBuiltinSchema(*dependency.Schema)
			dependency.Schema = &copy
		}
		schema.Dependencies[key] = dependency
	}
	return schema
}

func copyBuiltinSchemaSlice(schemas []spec.Schema) []spec.Schema {
	result := slices.Clone(schemas)
	for i, schema := range result {
		result[i] = copyBuiltinSchema(schema)
	}
	return result
}

func copyBuiltinSchemaOrBool(schema *spec.SchemaOrBool) *spec.SchemaOrBool {
	result := copyBuiltinPointer(schema)
	if result != nil && result.Schema != nil {
		copy := copyBuiltinSchema(*result.Schema)
		result.Schema = &copy
	}
	return result
}

func copyBuiltinPointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
