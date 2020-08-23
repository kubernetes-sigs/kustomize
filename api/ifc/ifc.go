// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package ifc holds miscellaneous interfaces used by kustomize.
package ifc

import (
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

// Validator provides functions to validate annotations and labels
type Validator interface {
	MakeAnnotationValidator() func(map[string]string) error
	MakeAnnotationNameValidator() func([]string) error
	MakeLabelValidator() func(map[string]string) error
	MakeLabelNameValidator() func([]string) error
	ValidateNamespace(string) []string
	ErrIfInvalidKey(string) error
	IsEnvVarName(k string) error
}

// KvLoader reads and validates KV pairs.
type KvLoader interface {
	Validator() Validator
	Load(args types.KvPairSources) (all []types.Pair, err error)
}

// Loader interface exposes methods to read bytes.
type Loader interface {
	// Root returns the root location for this Loader.
	Root() string
	// New returns Loader located at newRoot.
	New(newRoot string) (Loader, error)
	// Load returns the bytes read from the location or an error.
	Load(location string) ([]byte, error)
	// Cleanup cleans the loader
	Cleanup() error
}

// Kunstructured represents a Kubernetes Resource Model object.
type Kunstructured interface {
	// Several uses.
	Copy() Kunstructured

	// Used by Resource.Replace, which in turn is used in many places, e.g.
	//  - resource.Resource.Merge
	//  - resWrangler.appendReplaceOrMerge (AbsorbAll)
	//  - api.internal.k8sdeps.transformer.patch.conflictdetector
	GetAnnotations() map[string]string

	// Used by ResAccumulator and ReplacementTransformer.
	GetFieldValue(string) (interface{}, error)

	// Used by Resource.OrgId
	GetGvk() resid.Gvk

	// Used by resource.Factory.SliceFromBytes
	GetKind() string

	// Used by Resource.Replace
	GetLabels() map[string]string

	// Used by Resource.CurId and resource factory.
	GetName() string

	// Used by special case code in
	// ResMap.SubsetThatCouldBeReferencedByResource
	GetSlice(path string) ([]interface{}, error)

	// GetString returns the value of a string field.
	// Used by Resource.GetNamespace
	GetString(string) (string, error)

	// Several uses.
	Map() map[string]interface{}

	// Used by Resource.AsYAML and Resource.String
	MarshalJSON() ([]byte, error)

	// Used by resWrangler.Select
	MatchesAnnotationSelector(selector string) (bool, error)

	// Used by resWrangler.Select
	MatchesLabelSelector(selector string) (bool, error)

	// Used by Resource.Replace.
	SetAnnotations(map[string]string)

	// Used by PatchStrategicMergeTransformer.
	SetGvk(resid.Gvk)

	// Used by Resource.Replace and used to remove "validated by" labels.
	SetLabels(map[string]string)

	// Used by Resource.Replace.
	SetName(string)

	// Used by Resource.Replace.
	SetNamespace(string)

	// Needed, for now, by kyaml/filtersutil.ApplyToJSON.
	UnmarshalJSON([]byte) error
}

// KunstructuredFactory makes instances of Kunstructured.
type KunstructuredFactory interface {
	SliceFromBytes([]byte) ([]Kunstructured, error)
	FromMap(m map[string]interface{}) Kunstructured
	Hasher() KunstructuredHasher
	MakeConfigMap(kvLdr KvLoader, args *types.ConfigMapArgs) (Kunstructured, error)
	MakeSecret(kvLdr KvLoader, args *types.SecretArgs) (Kunstructured, error)
}

// KunstructuredHasher returns a hash of the argument
// or an error.
type KunstructuredHasher interface {
	Hash(Kunstructured) (string, error)
}

// See core.v1.SecretTypeOpaque
const SecretTypeOpaque = "Opaque"
