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

import "k8s.io/apimachinery/pkg/runtime/schema"

// selectByPath returns true if path `p` matches any `GroupVersionKind` and `Path` in `pathConfigs`.
func selectByPath(gvk schema.GroupVersionKind, p PathConfig, pathConfigs []PathConfig) bool {
	// If `pathConfigs` is nil, it is considered as a wildcard; always return true.
	if pathConfigs == nil {
		return true
	}
	if len(pathConfigs) > 0 {
		for _, pathConfig := range pathConfigs {
			if selectByGVK(gvk, pathConfig.GroupVersionKind) {
				// If p.Path or pathConfig.Path is nil, it is considered as a wildcard; always return true.
				if p.Path == nil || pathConfig.Path == nil {
					return true
				}
				if len(p.Path) != len(pathConfig.Path) {
					continue
				}
				allSegmentsMatch := true
				for i, segment := range p.Path {
					if segment != pathConfig.Path[i] {
						allSegmentsMatch = false
						break
					}
				}
				if allSegmentsMatch {
					return true
				}
			}
		}
	}
	return false
}
