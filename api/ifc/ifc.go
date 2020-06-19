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

// Kunstructured allows manipulation of k8s objects
// that do not have Golang structs.
type Kunstructured interface {
	Copy() Kunstructured
	GetAnnotations() map[string]string
	GetBool(path string) (bool, error)
	GetFieldValue(string) (interface{}, error)
	GetFloat64(path string) (float64, error)
	GetGvk() resid.Gvk
	GetInt64(path string) (int64, error)
	GetKind() string
	GetLabels() map[string]string
	GetMap(path string) (map[string]interface{}, error)
	GetName() string
	GetSlice(path string) ([]interface{}, error)
	GetString(string) (string, error)
	GetStringMap(path string) (map[string]string, error)
	GetStringSlice(string) ([]string, error)
	Map() map[string]interface{}
	MarshalJSON() ([]byte, error)
	MatchesAnnotationSelector(selector string) (bool, error)
	MatchesLabelSelector(selector string) (bool, error)
	Patch(Kunstructured) error
	SetAnnotations(map[string]string)
	SetGvk(resid.Gvk)
	SetLabels(map[string]string)
	SetMap(map[string]interface{})
	SetName(string)
	SetNamespace(string)
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
