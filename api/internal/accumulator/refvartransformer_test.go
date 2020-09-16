// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	resmaptest_test "sigs.k8s.io/kustomize/api/testutils/resmaptest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestRefVarTransformer(t *testing.T) {
	type given struct {
		varMap map[string]interface{}
		fs     []types.FieldSpec
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
		errMessage  string
	}{
		{
			description: "var replacement in map[string]",
			given: given{
				varMap: map[string]interface{}{
					"FOO": "replacementForFoo",
					"BAR": "replacementForBar",
					"BAZ": int64(5),
					"BOO": true,
				},
				fs: []types.FieldSpec{
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/map"},
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/slice"},
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/interface"},
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/num"},
				},
				res: resmaptest_test.NewRmBuilder(
					t, resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())).
					Add(map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm1",
						},
						"data": map[string]interface{}{
							"map": map[string]interface{}{
								"item1": "$(FOO)",
								"item2": "bla",
								"item3": "$(BAZ)",
								"item4": "$(BAZ)+$(BAZ)",
								"item5": "$(BOO)",
								"item6": "if $(BOO)",
								"item7": int64(2019),
							},
							"slice": []interface{}{
								"$(FOO)",
								"bla",
								"$(BAZ)",
								"$(BAZ)+$(BAZ)",
								"$(BOO)",
								"if $(BOO)",
							},
							"interface": "$(FOO)",
							"num":       int64(2019),
						}}).ResMap(),
			},
			expected: expected{
				res: resmaptest_test.NewRmBuilder(
					t, resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())).
					Add(map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm1",
						},
						"data": map[string]interface{}{
							"map": map[string]interface{}{
								"item1": "replacementForFoo",
								"item2": "bla",
								"item3": int64(5),
								"item4": "5+5",
								"item5": true,
								"item6": "if true",
								"item7": int64(2019),
							},
							"slice": []interface{}{
								"replacementForFoo",
								"bla",
								int64(5),
								"5+5",
								true,
								"if true",
							},
							"interface": "replacementForFoo",
							"num":       int64(2019),
						}}).ResMap(),
				unused: []string{"BAR"},
			},
		},
		{
			description: "var replacement panic in map[string]",
			given: given{
				varMap: map[string]interface{}{},
				fs: []types.FieldSpec{
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/slice"},
				},
				res: resmaptest_test.NewRmBuilder(
					t, resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())).
					Add(map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm1",
						},
						"data": map[string]interface{}{
							"slice": []interface{}{5}, // noticeably *not* a []string
						}}).ResMap(),
			},
			errMessage: `obj '{"apiVersion": "v1", "data": {"slice": [5]}, "kind": "ConfigMap", "metadata": {"name": "cm1"}}
' at path 'data/slice': invalid value type expect a string`,
		},
		{
			description: "var replacement in nil",
			given: given{
				varMap: map[string]interface{}{},
				fs: []types.FieldSpec{
					{Gvk: resid.Gvk{Version: "v1", Kind: "ConfigMap"}, Path: "data/nil"},
				},
				res: resmaptest_test.NewRmBuilder(
					t, resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())).
					Add(map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm1",
						},
						"data": map[string]interface{}{
							"nil": nil, // noticeably *not* a []string
						}}).ResMap(),
			},
			expected: expected{
				res: resmaptest_test.NewRmBuilder(
					t, resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())).
					Add(map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "cm1",
						},
						"data": map[string]interface{}{
							"nil": nil, // noticeably *not* a []string
						}}).ResMap(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// arrange
			tr := newRefVarTransformer(tc.given.varMap, tc.given.fs)

			// act
			err := tr.Transform(tc.given.res)

			// assert
			if tc.errMessage != "" {
				if err == nil {
					t.Fatalf("missing expected error %v", tc.errMessage)
				} else if err.Error() != tc.errMessage {
					t.Fatalf("actual error doesn't match expected error: \nACTUAL: %v\nEXPECTED: %v", err.Error(), tc.errMessage)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				a, e := tc.given.res, tc.expected.res
				if !reflect.DeepEqual(a, e) {
					err = e.ErrorIfNotEqualLists(a)
					t.Fatalf("actual doesn't match expected: \nACTUAL:\n%v\nEXPECTED:\n%v\nERR: %v", a, e, err)
				}
			}
		})
	}
}
