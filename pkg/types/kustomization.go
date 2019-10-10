// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package types holds struct definitions that should find a better home.
package types

import (
	"sigs.k8s.io/kustomize/v3/pkg/image"
)

const (
	KustomizationVersion = "kustomize.config.k8s.io/v1beta1"
	KustomizationKind    = "Kustomization"
)

// TypeMeta partially copies apimachinery/pkg/apis/meta/v1.TypeMeta
// No need for a direct dependence; the fields are stable.
type TypeMeta struct {
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// Kustomization holds the information needed to generate customized k8s api resources.
type Kustomization struct {
	TypeMeta `json:",inline" yaml:",inline"`

	//
	// Operators - what kustomize can do.
	//

	// NamePrefix will prefix the names of all resources mentioned in the kustomization
	// file including generated configmaps and secrets.
	NamePrefix string `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`

	// NameSuffix will suffix the names of all resources mentioned in the kustomization
	// file including generated configmaps and secrets.
	NameSuffix string `json:"nameSuffix,omitempty" yaml:"nameSuffix,omitempty"`

	// Namespace to add to all objects.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// CommonLabels to add to all objects and selectors.
	CommonLabels map[string]string `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`

	// CommonAnnotations to add to all objects.
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`

	// PatchesStrategicMerge specifies the relative path to a file
	// containing a strategic merge patch.  Format documented at
	// https://github.com/kubernetes/community/blob/master/contributors/devel/strategic-merge-patch.md
	// URLs and globs are not supported.
	PatchesStrategicMerge []PatchStrategicMerge `json:"patchesStrategicMerge,omitempty" yaml:"patchesStrategicMerge,omitempty"`

	// JSONPatches is a list of JSONPatch for applying JSON patch.
	// Format documented at https://tools.ietf.org/html/rfc6902
	// and http://jsonpatch.com
	PatchesJson6902 []PatchJson6902 `json:"patchesJson6902,omitempty" yaml:"patchesJson6902,omitempty"`

	// Patches is a list of patches, where each one can be either a
	// Strategic Merge Patch or a JSON patch.
	// Each patch can be applied to multiple target objects.
	Patches []Patch `json:"patches,omitempty" yaml:"patches,omitempty"`

	// Images is a list of (image name, new name, new tag or digest)
	// for changing image names, tags or digests. This can also be achieved with a
	// patch, but this operator is simpler to specify.
	Images []image.Image `json:"images,omitempty" yaml:"images,omitempty"`

	// Replicas is a list of {resourcename, count} that allows for simpler replica
	// specification. This can also be done with a patch.
	Replicas []Replica `json:"replicas,omitempty" yaml:"replicas,omitempty"`

	// Vars allow things modified by kustomize to be injected into a
	// kubernetes object specification. A var is a name (e.g. FOO) associated
	// with a field in a specific resource instance.  The field must
	// contain a value of type string/bool/int/float, and defaults to the name field
	// of the instance.  Any appearance of "$(FOO)" in the object
	// spec will be replaced at kustomize build time, after the final
	// value of the specified field has been determined.
	Vars []Var `json:"vars,omitempty" yaml:"vars,omitempty"`

	//
	// Operands - what kustomize operates on.
	//

	// Resources specifies relative paths to files holding YAML representations
	// of kubernetes API objects, or specifcations of other kustomizations
	// via relative paths, absolute paths, or URLs.
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	// Crds specifies relative paths to Custom Resource Definition files.
	// This allows custom resources to be recognized as operands, making
	// it possible to add them to the Resources list.
	// CRDs themselves are not modified.
	Crds []string `json:"crds,omitempty" yaml:"crds,omitempty"`

	// Deprecated.
	// Anything that would have been specified here should
	// be specified in the Resources field instead.
	Bases []string `json:"bases,omitempty" yaml:"bases,omitempty"`

	//
	// Generators (operators that create operands)
	//

	// ConfigMapGenerator is a list of configmaps to generate from
	// local data (one configMap per list item).
	// The resulting resource is a normal operand, subject to
	// name prefixing, patching, etc.  By default, the name of
	// the map will have a suffix hash generated from its contents.
	ConfigMapGenerator []ConfigMapArgs `json:"configMapGenerator,omitempty" yaml:"configMapGenerator,omitempty"`

	// SecretGenerator is a list of secrets to generate from
	// local data (one secret per list item).
	// The resulting resource is a normal operand, subject to
	// name prefixing, patching, etc.  By default, the name of
	// the map will have a suffix hash generated from its contents.
	SecretGenerator []SecretArgs `json:"secretGenerator,omitempty" yaml:"secretGenerator,omitempty"`

	// GeneratorOptions modify behavior of all ConfigMap and Secret generators.
	GeneratorOptions *GeneratorOptions `json:"generatorOptions,omitempty" yaml:"generatorOptions,omitempty"`

	// Configurations is a list of transformer configuration files
	Configurations []string `json:"configurations,omitempty" yaml:"configurations,omitempty"`

	// Generators is a list of files containing custom generators
	Generators []string `json:"generators,omitempty" yaml:"generators,omitempty"`

	// Transformers is a list of files containing transformers
	Transformers []string `json:"transformers,omitempty" yaml:"transformers,omitempty"`

	// Inventory appends an object that contains the record
	// of all other objects, which can be used in apply, prune and delete
	Inventory *Inventory `json:"inventory,omitempty" yaml:"inventory,omitempty"`
}

//go:generate stringer -type=GarbagePolicy
type GarbagePolicy int

const (
	GarbageIgnore GarbagePolicy = iota + 1
	GarbageCollect
)

// FixKustomizationPostUnmarshalling fixes things
// like empty fields that should not be empty, or
// moving content of deprecated fields to newer
// fields.
func (k *Kustomization) FixKustomizationPostUnmarshalling() {
	if k.APIVersion == "" {
		k.APIVersion = KustomizationVersion
	}
	if k.Kind == "" {
		k.Kind = KustomizationKind
	}
	// The EnvSource field is deprecated in favor of the list.
	for i, g := range k.ConfigMapGenerator {
		if g.EnvSource != "" {
			k.ConfigMapGenerator[i].EnvSources =
				append(g.EnvSources, g.EnvSource)
			k.ConfigMapGenerator[i].EnvSource = ""
		}
	}
	for i, g := range k.SecretGenerator {
		if g.EnvSource != "" {
			k.SecretGenerator[i].EnvSources =
				append(g.EnvSources, g.EnvSource)
			k.SecretGenerator[i].EnvSource = ""
		}
	}
	for _, b := range k.Bases {
		k.Resources = append(k.Resources, b)
	}
	k.Bases = nil
}

func (k *Kustomization) EnforceFields() []string {
	var errs []string
	if k.APIVersion != "" && k.APIVersion != KustomizationVersion {
		errs = append(errs, "apiVersion should be "+KustomizationVersion)
	}
	if k.Kind != "" && k.Kind != KustomizationKind {
		errs = append(errs, "kind should be "+KustomizationKind)
	}
	return errs
}

// PluginConfig holds plugin configuration.
type PluginConfig struct {
	// DirectoryPath is an absolute path to a
	// directory containing kustomize plugins.
	// This directory may contain subdirectories
	// further categorizing plugins.
	DirectoryPath string

	// Enabled is true if plugins are enabled.
	Enabled bool
}

// Pair is a key value pair.
type Pair struct {
	Key   string
	Value string
}

type PluginType string

func (p PluginType) IsUndefined() bool {
	return p == PluginType("")
}

// KVSource represents a KV plugin backend.
type KVSource struct {
	PluginType PluginType `json:"pluginType,omitempty" yaml:"pluginType,omitempty"`
	Name       string     `json:"name,omitempty" yaml:"name,omitempty"`
	Args       []string   `json:"args,omitempty" yaml:"args,omitempty"`
}

type Inventory struct {
	Type      string   `json:"type,omitempty" yaml:"type,omitempty"`
	ConfigMap NameArgs `json:"configMap,omitempty" yaml:"configMap,omitempty"`
}

type NameArgs struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// PatchJson6902 represents a json patch for an object
// with format documented https://tools.ietf.org/html/rfc6902.
type PatchJson6902 struct {
	// PatchTarget refers to a Kubernetes object that the json patch will be
	// applied to. It must refer to a Kubernetes resource under the
	// purview of this kustomization. PatchTarget should use the
	// raw name of the object (the name specified in its YAML,
	// before addition of a namePrefix and a nameSuffix).
	Target *PatchTarget `json:"target" yaml:"target"`

	// relative file path for a json patch file inside a kustomization
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// inline patch string
	Patch string `json:"patch,omitempty" yaml:"patch,omitempty"`
}

// Patch represent either a Strategic Merge Patch or a JSON patch
// and its targets.
// The content of the patch can either be from a file
// or from an inline string.
type Patch struct {
	// Path is a relative file path to the patch file.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Patch is the content of a patch.
	Patch string `json:"patch,omitempty" yaml:"patch,omitempty"`

	// Target points to the resources that the patch is applied to
	Target *Selector `json:"target,omitempty" yaml:"target,omitempty"`
}
