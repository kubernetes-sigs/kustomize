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

package config

import (
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

type gvkSlice []gvk.Gvk

func (s gvkSlice) Len() int      { return len(s) }
func (s gvkSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s gvkSlice) Less(i, j int) bool {
	return s[i].IsLessThan(s[j])
}

// mergeAll merges the argument into this, returning the result.
// Items already present are ignored.
func (s gvkSlice) merge(incoming gvkSlice) (result gvkSlice) {
	result = s
	for _, x := range incoming {
		i := s.index(x)
		if i > -1 {
			// It's already there.
			continue
		}
		result = append(result, x)
	}
	return
}

func (s gvkSlice) index(gvk gvk.Gvk) int {
	for i, x := range s {
		if x.IsSelected(&gvk) {
			return i
		}
	}
	return -1
}
