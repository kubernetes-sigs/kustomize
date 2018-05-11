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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PathConfig contains the configuration of a field, including the GVK it ties to,
// path to the field, etc.
type PathConfig struct {
	// If true, it will create the path if it is not found.
	CreateIfNotPresent bool
	// The GVK that this path tied to.
	// If unset, it applied to any GVK
	// If some fields are set, it applies to all matching GVK.
	GroupVersionKind *schema.GroupVersionKind
	// Path to the field that will be munged.
	Path []string
}

// referencePathConfig contains the configuration of a field that references
// the name of another resource whose GroupVersionKind is specified in referencedGVK.
// e.g. pod.spec.template.volumes.configMap.name references the name of a configmap
// Its corresponding referencePathConfig will look like:
//
//	referencePathConfig{
//	referencedGVK: schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
//	pathConfigs: []PathConfig{
//		{
//			GroupVersionKind: &schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
//			Path:             []string{"spec", "volumes", "configMap", "name"},
//		},
//	}
type referencePathConfig struct {
	// referencedGVK is the GroupVersionKind that is referenced by
	// the PathConfig's GVK in the path of PathConfig.Path.
	referencedGVK schema.GroupVersionKind
	// PathConfig is the GVK that is referencing the referencedGVK object's name.
	pathConfigs []PathConfig
}
