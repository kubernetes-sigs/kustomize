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
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

func TestVarRef(t *testing.T) {
	type given struct {
		varMap map[string]string
		fs     []config.FieldSpec
		res    resmap.ResMap
	}
	type expected struct {
		res    resmap.ResMap
		unused []string
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			description: "var replacement in map[string]",
			given: given{
				varMap: map[string]string{
					"FOO": "replacementForFoo",
					"BAR": "replacementForBar",
				},
				fs: []config.FieldSpec{
					{Gvk: cmap, Path: "data"},
				},
				res: resmap.FromMap(map[resid.ResId]*resource.Resource{
					resid.NewResId(cmap, "cm1"): rf.FromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "cm1",
							},
							"data": map[string]interface{}{
								"item1": "$(FOO)",
								"item2": "bla",
							},
						}),
				}),
			},
			expected: expected{
				res: resmap.FromMap(map[resid.ResId]*resource.Resource{
					resid.NewResId(cmap, "cm1"): rf.FromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "cm1",
							},
							"data": map[string]interface{}{
								"item1": "replacementForFoo",
								"item2": "bla",
							},
						}),
				}),
				unused: []string{"BAR"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// arrange
			tr := NewRefVarTransformer(tc.given.varMap, tc.given.fs)

			// act
			err := tr.Transform(tc.given.res)

			// assert
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			a, e := tc.given.res, tc.expected.res
			if !reflect.DeepEqual(a, e) {
				err = e.ErrorIfNotEqual(a)
				t.Fatalf("actual doesn't match expected: \nACTUAL:\n%v\nEXPECTED:\n%v\nERR: %v", a, e, err)
			}

		})
	}
}
