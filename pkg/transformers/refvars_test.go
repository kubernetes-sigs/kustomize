package transformers

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
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
				res: resmap.ResMap{
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
				},
			},
			expected: expected{
				res: resmap.ResMap{
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
				},
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
