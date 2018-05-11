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

package types

// Kustomization holds the information needed to generate customized k8s api resources.
type Kustomization struct {
	// NamePrefix will prefix the names of all resources mentioned in the kustomization
	// file including generated configmaps and secrets.
	NamePrefix string `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`

	// Labels to add to all objects and selectors.
	// These labels would also be used to form the selector for apply --prune
	// Named differently than “labels” to avoid confusion with metadata for
	// this object
	CommonLabels map[string]string `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`

	// Annotations to add to all objects.
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`

	// Each entry should be either a path to a file with a name matching the value of
	// constants.KustomizationFileName, or a path to a directory containing a file with that name.
	Bases []string `json:"bases,omitempty" yaml:"bases,omitempty"`

	// Resources specifies the relative paths within the package.
	// It could be any format that kubectl -f allows, i.e. files, directories,
	// URLs and globs.
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	// An Patch entry is very similar to an Resource entry.
	// It specifies the relative paths within the package, and could be any
	// format that kubectl -f allows.
	// It should be able to be merged by Strategic Merge Patch on top of its
	// corresponding base resource.
	Patches []string `json:"patches,omitempty" yaml:"patches,omitempty"`

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
}

// ConfigMapArg contains the metadata of how to generate a configmap.
type ConfigMapArgs struct {
	// Name of the configmap.
	// The full name should be Kustomization.NamePrefix + Configmap.Name +
	// hash(content of configmap).
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Behavior of configmap, must be one of create, merge and replace
	// 'create': create a new one;
	// 'replace': replace the existing one;
	// 'merge': merge the existing one.
	Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`

	// DataSources for configmap.
	DataSources `json:",inline,omitempty" yaml:",inline,omitempty"`
}

// SecretGenerator contains the metadata of how to generate a secret.
type SecretArgs struct {
	// Name of the secret.
	// The full name should be Kustomization.NamePrefix + SecretGenerator.Name +
	// hash(content of secret).
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Behavior of secretGenerator, must be one of create, merge and replace
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

	// Map of keys to commands to generate the values
	Commands map[string]string `json:",commands,omitempty" yaml:",inline,omitempty"`
}

// DataSources contains some generic sources for configmap or secret.
// Only one field can be set.
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
