// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// globalSchema contains global state information about the openapi
var globalSchema openapiData

// openapiData contains the parsed openapi state.  this is in a struct rather than
// a list of vars so that it can be reset from tests.
type openapiData struct {
	setup                sync.Once
	schema               spec.Schema
	schemaByResourceType map[yaml.TypeMeta]*spec.Schema
	noUseBuiltInSchema   bool
}

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
	rs, found := globalSchema.schemaByResourceType[t]
	if !found {
		return nil
	}
	return &ResourceSchema{Schema: rs}
}

// AddSchema parses s, and adds definitions from s to the global schema.
func AddSchema(s []byte) (*spec.Schema, error) {
	return parse(s)
}

// AddDefinitions adds the definitions to the global schema.
func AddDefinitions(definitions spec.Definitions) {
	// initialize values if they have not yet been set
	if globalSchema.schemaByResourceType == nil {
		globalSchema.schemaByResourceType = map[yaml.TypeMeta]*spec.Schema{}
	}
	if globalSchema.schema.Definitions == nil {
		globalSchema.schema.Definitions = spec.Definitions{}
	}

	// index the schema definitions so we can lookup them up for Resources
	for k := range definitions {
		// index by GVK, if no GVK is found then it is the schema for a subfield
		// of a Resource
		d := definitions[k]

		// copy definitions to the schema
		globalSchema.schema.Definitions[k] = d
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
		globalSchema.schemaByResourceType[yaml.TypeMeta{Kind: m[kindKey].(string), APIVersion: apiVersion}] = &d
	}
}

// Resolve resolves the reference against the global schema
func Resolve(ref *spec.Ref) (*spec.Schema, error) {
	return resolve(Schema(), ref)
}

// Schema returns the global schema
func Schema() *spec.Schema {
	return rootSchema()
}

// GetSchema parses s into a ResourceSchema, resolving References within the
// global schema.
func GetSchema(s string) (*ResourceSchema, error) {
	var sc spec.Schema
	if err := sc.UnmarshalJSON([]byte(s)); err != nil {
		return nil, errors.Wrap(err)
	}
	if sc.Ref.String() != "" {
		r, err := Resolve(&sc.Ref)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		sc = *r
	}

	return &ResourceSchema{Schema: &sc}, nil
}

// SuppressBuiltInSchemaUse can be called to prevent using the built-in Kubernetes
// schema as part of the global schema.
// Must be called before the schema is used.
func SuppressBuiltInSchemaUse() {
	globalSchema.noUseBuiltInSchema = false
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
		sc, e := Resolve(&s.Ref)
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
		sc, e := Resolve(&s.Ref)
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
	globalSchema.setup.Do(func() {
		if globalSchema.noUseBuiltInSchema {
			// don't parse the built in schema
			return
		}

		// parse the swagger, this should never fail
		if _, err := parse(MustAsset(openAPIAssetName)); err != nil {
			// this should never happen
			panic(err)
		}
	})
}

// parse parses and indexes a single json schema
func parse(b []byte) (*spec.Schema, error) {
	var sc spec.Schema

	if err := sc.UnmarshalJSON(b); err != nil {
		return nil, errors.Wrap(err)
	}
	AddDefinitions(sc.Definitions)
	return &sc, nil
}

func resolve(root interface{}, ref *spec.Ref) (*spec.Schema, error) {
	res, _, err := ref.GetPointer().Get(root)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	switch sch := res.(type) {
	case spec.Schema:
		return &sch, nil
	case *spec.Schema:
		return sch, nil
	case map[string]interface{}:
		b, err := json.Marshal(sch)
		if err != nil {
			return nil, err
		}
		newSch := new(spec.Schema)
		if err = json.Unmarshal(b, newSch); err != nil {
			return nil, err
		}
		return newSch, nil
	default:
		return nil, errors.Wrap(fmt.Errorf("unknown type for the resolved reference"))
	}
}

func rootSchema() *spec.Schema {
	initSchema()
	return &globalSchema.schema
}
