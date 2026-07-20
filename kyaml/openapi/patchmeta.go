// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// pmField is one field of a definition in the precomputed patch-metadata
// table (zz_generated_patchmeta.go), restricted to exactly what the
// strategic-merge walker consumes from the full OpenAPI schema.
type pmField struct {
	Strategy   string   // x-kubernetes-patch-strategy
	MergeKey   string   // x-kubernetes-patch-merge-key
	MergeKeys  []string // x-kubernetes-list-map-keys
	Ref        string   // object-field / map-value definition
	ElementRef string   // array-element definition
	IsArray    bool
	IsMap      bool
}

// canUsePrecomputedPatchMeta reports whether lookups may be served from the
// precomputed table: only in the pristine default state, i.e. no custom
// schema, no explicit builtin version, builtin schema not suppressed, and no
// definitions added or parsed into the global schema. Any customization
// falls back to the full-schema path so existing behavior is preserved
// exactly.
func canUsePrecomputedPatchMeta() bool {
	schemaLock.RLock()
	defer schemaLock.RUnlock()
	return customSchema == nil &&
		kubernetesOpenAPIVersion == "" &&
		!globalSchema.noUseBuiltInSchema &&
		!globalSchema.schemaInit &&
		len(globalSchema.schema.Definitions) == 0
}

// precomputedSchemaForResourceType returns a table-backed ResourceSchema for
// t. The second return is false when t is not a builtin type known to the
// table, in which case the caller must fall back to the full-schema path.
func precomputedSchemaForResourceType(t yaml.TypeMeta) (*ResourceSchema, bool) {
	def, known := precomputedGVKToDef[t]
	if !known {
		return nil, false
	}
	return pmDefNode(def), true
}

// pmDefNode returns a ResourceSchema backed by the named definition in the
// precomputed table. An empty or unknown name yields an empty node: known
// type, no strategic-merge metadata anywhere below (equivalent to a schema
// without patch extensions, which the walker treats like no schema at all).
func pmDefNode(def string) *ResourceSchema {
	s := &spec.Schema{}
	if _, ok := precomputedPatchDefs[def]; ok {
		// mark non-empty so IsMissingOrNull stays false, mirroring a real
		// definition node
		s.Properties = map[string]spec.Schema{}
	}
	return &ResourceSchema{Schema: s, pmKnown: true, pmDef: def}
}

// pmFieldNode synthesizes the schema node for a field carrying patch
// extensions and/or array/map structure, mirroring what the full schema
// would expose on the same property.
func pmFieldNode(f pmField) *ResourceSchema {
	s := &spec.Schema{}
	if f.IsArray {
		s.Type = spec.StringOrArray{"array"}
	}
	if f.Strategy != "" {
		s.AddExtension(kubernetesPatchStrategyExtensionKey, f.Strategy)
	}
	if f.MergeKey != "" {
		s.AddExtension(kubernetesMergeKeyExtensionKey, f.MergeKey)
	}
	if len(f.MergeKeys) > 0 {
		keys := make([]interface{}, len(f.MergeKeys))
		for i := range f.MergeKeys {
			keys[i] = f.MergeKeys[i]
		}
		s.AddExtension(kubernetesMergeKeyMapList, keys)
	}
	fc := f
	return &ResourceSchema{Schema: s, pmKnown: true, pmField: &fc}
}

// pmBacked reports whether rs is served from the precomputed table.
func (rs *ResourceSchema) pmBacked() bool {
	return rs != nil && rs.pmKnown
}

// pmFieldLookup implements Field for table-backed nodes.
func (rs *ResourceSchema) pmFieldLookup(field string) *ResourceSchema {
	if rs.pmField != nil {
		if rs.pmField.IsMap {
			// map field: any key resolves to the value type
			return pmDefNode(rs.pmField.Ref)
		}
		// array/leaf field nodes have no named fields
		return nil
	}
	fields, ok := precomputedPatchDefs[rs.pmDef]
	if !ok {
		return nil
	}
	f, ok := fields[field]
	if !ok {
		return nil
	}
	if f.IsArray || f.IsMap || f.Strategy != "" {
		return pmFieldNode(f)
	}
	// plain object field: descend straight into its definition
	return pmDefNode(f.Ref)
}

// pmElements implements Elements for table-backed nodes.
func (rs *ResourceSchema) pmElements() *ResourceSchema {
	if rs.pmField == nil || !rs.pmField.IsArray {
		return nil
	}
	return pmDefNode(rs.pmField.ElementRef)
}
