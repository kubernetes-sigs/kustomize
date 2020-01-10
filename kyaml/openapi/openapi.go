// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"sync"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var setup sync.Once
var schema spec.Schema
var schemaByResourceType map[yaml.TypeMeta]*spec.Schema

// ResourceSchema wraps the OpenAPI Schema.
type ResourceSchema struct {
	// Schema is the OpenAPI schema for a Resource or field
	Schema *spec.Schema
}

// SchemaForResourceType returns the Schema for the given Resource
// TODO(pwittrock): create a version of this function that will return a schema
// which can be used for duck-typed Resources -- e.g. contains common fields such
// as metadata, replicas and spec.template.spec
func SchemaForResourceType(t yaml.TypeMeta) *ResourceSchema {
	initSchema()
	rs, found := schemaByResourceType[t]
	if !found {
		return nil
	}
	return &ResourceSchema{Schema: rs}
}

// Elements returns the Schema for the elements of an array.
func (r *ResourceSchema) Elements() *ResourceSchema {
	// load the schema from swagger.json
	initSchema()

	if len(r.Schema.Type) != 1 || r.Schema.Type[0] != "array" {
		// either not an array, or array has multiple types
		return nil
	}
	s := *r.Schema.Items.Schema
	for s.Ref.String() != "" {
		sc, e := spec.ResolveRef(rootSchema(), &s.Ref)
		if e != nil {
			return nil
		}
		s = *sc
	}
	return &ResourceSchema{Schema: &s}
}

const Elements = "[]"

// Lookup calls either Field or Elements for each item in the path.
// If the path item is "[]", then Elements is called, otherwise
// Field is called.
// If any Field or Elements call returns nil, then Lookup returns
// nil immediately.
func (r *ResourceSchema) Lookup(path ...string) *ResourceSchema {
	s := r
	for _, p := range path {
		if s == nil {
			break
		}
		if p == Elements {
			s = s.Elements()
			continue
		}
		s = s.Field(p)
	}
	return s
}

// Field returns the Schema for a field.
func (r *ResourceSchema) Field(field string) *ResourceSchema {
	// load the schema from swagger.json
	initSchema()

	// locate the Schema
	s, found := r.Schema.Properties[field]
	switch {
	case found:
		// no-op, continue with s as the schema
	case r.Schema.AdditionalProperties != nil && r.Schema.AdditionalProperties.Schema != nil:
		// map field type -- use Schema of the value
		// (the key doesn't matter, they all have the same value type)
		s = *r.Schema.AdditionalProperties.Schema
	default:
		// no Schema found from either swagger.json or line comments
		return nil
	}

	// resolve the reference to the Schema if the Schema has one
	for s.Ref.String() != "" {
		sc, e := spec.ResolveRef(rootSchema(), &s.Ref)
		if e != nil {
			return nil
		}
		s = *sc
	}

	// return the merged Schema
	return &ResourceSchema{Schema: &s}
}

// PatchStrategyAndKey returns the patch strategy and merge key extensions
func (r *ResourceSchema) PatchStrategyAndKey() (string, string) {
	ps, found := r.Schema.Extensions[kubernetesPatchStrategyExtensionKey]
	if !found {
		// merge key and patch strategy must appear together
		return "", ""
	}

	mk, found := r.Schema.Extensions[kubernetesMergeKeyExtensionKey]
	if !found {
		// merge key and patch strategy must appear together
		return "", ""
	}
	return ps.(string), mk.(string)
}

const (
	// openAPIAssetName is the name of the asset containing the statically compiled in
	// OpenAPI definitions for Kubernetes built-in types
	openAPIAssetName = "openapi/swagger.json"

	// kubernetesGVKExtensionKey is the key to lookup the kubernetes group version kind extension
	// -- the extension is an array of objects containing a gvk
	kubernetesGVKExtensionKey = "x-kubernetes-group-version-kind"
	// kubernetesMergeKeyExtensionKey is the key to lookup the kubernetes merge key extension
	// -- the extension is a string
	kubernetesMergeKeyExtensionKey = "x-kubernetes-patch-merge-key"
	// kubernetesPatchStrategyExtensionKey is the key to lookup the kubernetes patch strategy
	// extension -- the extension is a string
	kubernetesPatchStrategyExtensionKey = "x-kubernetes-patch-strategy"

	// groupKey is the key to lookup the group from the GVK extension
	groupKey = "group"
	// versionKey is the key to lookup the version from the GVK extension
	versionKey = "version"
	// kindKey is the the to lookup the kind from the GVK extension
	kindKey = "kind"
)

// initSchema parses the json schema
func initSchema() {
	setup.Do(func() {
		// initialize the map
		schemaByResourceType = map[yaml.TypeMeta]*spec.Schema{}

		// parse the swagger, this should never fail
		parse(MustAsset(openAPIAssetName))

		// TODO(pwittrock): add support for parsing additional schemas from
		// environment variables, files or other sources
	})
}

// parse parses and indexes a single json schema
func parse(b []byte) {
	if err := schema.UnmarshalJSON(b); err != nil {
		panic(err)
	}
	// index the schema definitions so we can lookup them up for Resources
	for k := range schema.Definitions {
		// index by GVK, if no GVK is found then it is the schema for a subfield
		// of a Resource
		d := schema.Definitions[k]
		gvk, found := d.VendorExtensible.Extensions[kubernetesGVKExtensionKey]
		if !found {
			continue
		}
		// cast the extension to a []map[string]string
		exts, ok := gvk.([]interface{})
		if !ok || len(exts) != 1 {
			continue
		}
		m, ok := exts[0].(map[string]interface{})
		if !ok {
			continue
		}

		// build the index key and save it
		g := m[groupKey].(string)
		apiVersion := m[versionKey].(string)
		if g != "" {
			apiVersion = g + "/" + apiVersion
		}
		schemaByResourceType[yaml.TypeMeta{Kind: m[kindKey].(string), APIVersion: apiVersion}] = &d
	}
}

func rootSchema() *spec.Schema {
	initSchema()
	return &schema
}
