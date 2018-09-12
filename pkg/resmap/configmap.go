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

package resmap

import (
	"sigs.k8s.io/kustomize/pkg/configmapandsecret"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

// NewResMapFromConfigMapArgs returns a Resource slice given
// a configmap metadata slice from kustomization file.
func NewResMapFromConfigMapArgs(
	f *configmapandsecret.ConfigMapFactory,
	cmArgsList []types.ConfigMapArgs) (ResMap, error) {
	var allResources []*resource.Resource
	for _, cmArgs := range cmArgsList {
		if cmArgs.Behavior == "" {
			cmArgs.Behavior = "create"
		}
		cm, err := f.MakeConfigMap(&cmArgs)
		if err != nil {
			return nil, err
		}
		res, err := resource.NewResourceWithBehavior(
			cm, resource.NewGenerationBehavior(cmArgs.Behavior))
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, res)
	}
	return newResMapFromResourceSlice(allResources)
}
