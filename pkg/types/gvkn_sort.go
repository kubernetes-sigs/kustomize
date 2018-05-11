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

package types

import (
	"sort"
)

// ByGVKN implements the sort interface.
type ByGVKN []GroupVersionKindName

var _ sort.Interface = ByGVKN{}

func (a ByGVKN) Len() int      { return len(a) }
func (a ByGVKN) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByGVKN) Less(i, j int) bool {
	if a[i].GVK.String() != a[j].GVK.String() {
		return a[i].GVK.String() < a[j].GVK.String()
	}
	return a[i].Name < a[j].Name
}
