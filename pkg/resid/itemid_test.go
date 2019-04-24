/*
Copyright 2019 The Kubernetes Authors.

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

package resid

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
)

var itemIds = []ItemId{
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Group: "g", Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Version: "v", Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{Kind: "k"},
		Name:      "nm",
	},
	{
		Namespace: "ns",
		Gvk:       gvk.Gvk{},
		Name:      "nm",
	},
	{
		Gvk:  gvk.Gvk{},
		Name: "nm",
	},
	{
		Gvk:  gvk.Gvk{},
		Name: "nm",
	},
	{
		Gvk: gvk.Gvk{},
	},
}

func TestItemIds(t *testing.T) {
	for _, item := range itemIds {
		newItem := FromString(item.String())
		if newItem != item {
			t.Fatalf("Actual: %v,  Expected: '%s'", newItem, item)
		}
	}
}
