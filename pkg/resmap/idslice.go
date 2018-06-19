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
	"sort"

	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IdSlice implements the sort interface.
type IdSlice []resource.ResId

var _ sort.Interface = IdSlice{}

func (a IdSlice) Len() int      { return len(a) }
func (a IdSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a IdSlice) Less(i, j int) bool {
	if a[i].Gvk().String() != a[j].Gvk().String() {
		return gvkLess(a[i].Gvk(), a[j].Gvk())
	}
	return a[i].Name() < a[j].Name()
}

func gvkLess(i, j schema.GroupVersionKind) bool {
	if i.Kind == "Namespace" {
		return true
	} else if j.Kind == "Namespace" {
		return false
	} else {
		return i.String() < j.String()
	}
}
