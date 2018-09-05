/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package types holds struct definitions that should find a better home.
package types

import (
	"github.com/kubernetes-sigs/kustomize/pkg/patch"
)

// Kustomization holds the information needed to generate customized k8s api resources.
type Kustomization struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
	// +optional
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`

	// NamePrefix will prefix the names of all resources mentioned in the kustomization
	// file including generated configmaps and secrets.
	NamePrefix string `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`

	// Namespace to add to all objects.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Labels to add to all objects and selectors.
	// These labels would also be used to form the selector for apply --prune
	// Named differently than “labels” to avoid confusion with metadata for
	// this object
	CommonLabels map[string]string `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`

	// Annotations to add to all objects.
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`

	// Each entry should be either a path to a directory containing kustomization.yaml
	// Or a repo URL pointing to a remote directory containing kustomization.yaml
	// The repo URL should follow hashicorp/go-getter URL format
	// https://github.com/hashicorp/go-getter#url-format
	Bases []string `json:"bases,omitempty" yaml:"bases,omitempty"`

	// Resources specifies the relative paths for resource files within the package.
	// URLs and globs are not supported
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	// Crds specifies relative paths to custom resource definition files.
	Crds []string `json:"crds,omitempty" yaml:"crds,omitempty"`

	// An Patch entry is very similar to an Resource entry.
	// It specifies the relative paths for patch files within the package.
	// URLs and globs are not supported.
	// The patch files should be Stategic Merge Patch, the default patching behavior for kubectl.
	// https://github.com/kubernetes/community/blob/master/contributors/devel/strategic-merge-patch.md
	Patches               []string                    `json:"patches,omitempty" yaml:"patches,omitempty"`
	PatchesStrategicMerge []patch.PatchStrategicMerge `json:"patchesStrategicMerge,omitempty" yaml:"patchesStrategicMerge,omitempty"`

	// JSONPatches is a list of JSONPatch for applying JSON patch.
	// The JSON patch is documented at https://tools.ietf.org/html/rfc6902
	// and http://jsonpatch.com/.
	PatchesJson6902 []patch.PatchJson6902 `json:"patchesJson6902,omitempty" yaml:"patchesJson6902,omitempty"`

	// List of configmaps to generate from configuration sources.
	// Base/overlay concept doesn't apply to this field.
	// If a configmap want to have a base and an overlay, it should go to Bases
	// and Overlays fields.
	ConfigMapGenerator []ConfigMapArgs `json:"configMapGenerator,omitempty" yaml:"configMapGenerator,omitempty"`

	// List of secrets to generate from secret commands.
	// Base/overlay concept doesn't apply to this field.
	// If a secret want to have a base and an overlay, it should go to Bases and
	// Overlays fields.
	SecretGenerator []SecretArgs `json:"secretGenerator,omitempty" yaml:"secretGenerator,omitempty"`

	// Variables which will be substituted at runtime
	Vars []Var `json:"vars,omitempty" yaml:"vars,omitempty"`

	// If set to true, all images need to have tags
	RequireTag bool `json:"requireTag,omitempty" yaml:"requireTag,omitempty"`

	// ImageTags is a list of ImageTag for changing image tags
	ImageTags []ImageTag `json:"imageTags,omitempty" yaml:"imageTags,omitempty"`
}

// ConfigMapArgs contains the metadata of how to generate a configmap.
type ConfigMapArgs struct {
	// Name of the configmap.
	// The full name should be Kustomization.NamePrefix + Configmap.Name +
	// hash(content of configmap).
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// behavior of configmap, must be one of create, merge and replace
	// 'create': create a new one;
	// 'replace': replace the existing one;
	// 'merge': merge the existing one.
	Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`

	// DataSources for configmap.
	DataSources `json:",inline,omitempty" yaml:",inline,omitempty"`
}

// SecretArgs contains the metadata of how to generate a secret.
type SecretArgs struct {
	// Name of the secret.
	// The full name should be Kustomization.NamePrefix + SecretGenerator.Name +
	// hash(content of secret).
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// behavior of secretGenerator, must be one of create, merge and replace
	// 'create': create a new one;
	// 'replace': replace the existing one;
	// 'merge': merge the existing one.
	Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`

	// Type of the secret.
	//
	// This is the same field as the secret type field in v1/Secret:
	// It can be "Opaque" (default), or "kubernetes.io/tls".
	//
	// If type is "kubernetes.io/tls", then "Commands" must have exactly two
	// keys: "tls.key" and "tls.crt"
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// CommandSources for secret.
	CommandSources `json:",inline,omitempty" yaml:",inline,omitempty"`

	// TimeoutSeconds specifies the timeout for commands.
	TimeoutSeconds *int64 `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty"`
}

// CommandSources contains some generic sources for secrets.
type CommandSources struct {
	// Map of keys to commands to generate the values
	Commands map[string]string `json:"commands,omitempty" yaml:"commands,omitempty"`
	// EnvCommand to output lines of key=val pairs to create a secret.
	// i.e. a Docker .env file or a .ini file.
	EnvCommand string `json:"envCommand,omitempty" yaml:"envCommand,omitempty"`
}

// DataSources contains some generic sources for configmaps.
type DataSources struct {
	// LiteralSources is a list of literal sources.
	// Each literal source should be a key and literal value,
	// e.g. `somekey=somevalue`
	// It will be similar to kubectl create configmap|secret --from-literal
	LiteralSources []string `json:"literals,omitempty" yaml:"literals,omitempty"`

	// FileSources is a list of file sources.
	// Each file source can be specified using its file path, in which case file
	// basename will be used as configmap key, or optionally with a key and file
	// path, in which case the given key will be used.
	// Specifying a directory will iterate each named file in the directory
	// whose basename is a valid configmap key.
	// It will be similar to kubectl create configmap|secret --from-file
	FileSources []string `json:"files,omitempty" yaml:"files,omitempty"`

	// EnvSource format should be a path to a file to read lines of key=val
	// pairs to create a configmap.
	// i.e. a Docker .env file or a .ini file.
	EnvSource string `json:"env,omitempty" yaml:"env,omitempty"`
}

// ImageTag contains an image and a new tag, which will replace the original tag.
type ImageTag struct {
	// Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// NewTag is the value to use in replacing the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`
}
