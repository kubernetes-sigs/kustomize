/*
Copyright 2018 The Kubernetes Authors.

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

package transformers

import (
	"sigs.k8s.io/kustomize/pkg/gvk"
)

// PathConfig contains the configuration of a field, including the gvk it ties to,
// path to the field, etc.
type PathConfig struct {
	// If true, it will create the path if it is not found.
	CreateIfNotPresent bool
	// The gvk that this path tied to.
	// If unset, it applied to any gvk
	// If some fields are set, it applies to all matching gvk.
	GroupVersionKind *gvk.Gvk
	// Path to the field that will be munged.
	Path []string
}

// ReferencePathConfig contains the configuration of a field that references
// the name of another resource whose GroupVersionKind is specified in referencedGVK.
// e.g. pod.spec.template.volumes.configMap.name references the name of a configmap
// Its corresponding referencePathConfig will look like:
//
//	ReferencePathConfig{
//	referencedGVK: schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
//	pathConfigs: []PathConfig{
//		{
//			GroupVersionKind: &schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
//			Path:             []string{"spec", "volumes", "configMap", "name"},
//		},
//	}
type ReferencePathConfig struct {
	// referencedGVK is the GroupVersionKind that is referenced by
	// the PathConfig's gvk in the path of PathConfig.Path.
	referencedGVK gvk.Gvk
	// PathConfig is the gvk that is referencing the referencedGVK object's name.
	pathConfigs []PathConfig
}

// NewReferencePathConfig creates a new ReferencePathConfig object
func NewReferencePathConfig(k gvk.Gvk, pathconfigs []PathConfig) ReferencePathConfig {
	return ReferencePathConfig{
		referencedGVK: k,
		pathConfigs:   pathconfigs,
	}
}

// GVK returns the Group version kind of a Reference PathConfig
func (r ReferencePathConfig) GVK() string {
	return r.referencedGVK.String()
}
