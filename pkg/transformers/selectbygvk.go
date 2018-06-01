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

// selectByGVK returns true if `selector` selects `in`; otherwise, false.
// If `selector` and `in` are the same, return true.
// If `selector` is nil, it is considered as a wildcard and always return true.
// e.g. selector <Group: "", Version: "", Kind: "Deployment"> CAN select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployment">.
// selector <Group: "apps", Version: "", Kind: "Deployment"> CANNOT select
// <Group: "extensions", Version: "v1beta1", Kind: "Deployment">.
func selectByGVK(in schema.GroupVersionKind, selector *schema.GroupVersionKind) bool {
	if selector == nil {
		return true
	}
	if len(selector.Group) > 0 {
		if in.Group != selector.Group {
			return false
		}
	}
	if len(selector.Version) > 0 {
		if in.Version != selector.Version {
			return false
		}
	}
	if len(selector.Kind) > 0 {
		if in.Kind != selector.Kind {
			return false
		}
	}
	return true
}
