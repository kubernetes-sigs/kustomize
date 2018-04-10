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

package resource

import (
	"k8s.io/kubectl/pkg/loader"
)

//  NewFromResources returns a ResourceCollection given a resource path slice from manifest file.
func NewFromResources(loader loader.Loader, paths []string) (ResourceCollection, error) {
	allResources := []ResourceCollection{}
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, err
		}

		res, err := decodeToResourceCollection(content)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, res)
	}
	return Merge(allResources...)
}

//  NewFromPatches returns a slice of Resources given a patch path slice from manifest file.
func NewFromPatches(loader loader.Loader, paths []string) ([]*Resource, error) {
	allResources := []*Resource{}
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, err
		}

		res, err := decode(content)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, res...)
	}
	return allResources, nil
}
