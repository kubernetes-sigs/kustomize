package refvar_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/filters/refvar"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var makeMf = func(theMap map[string]interface{}) MappingFunc {
	ignored := make(map[string]int)
	return MakePrimitiveReplacer(ignored, theMap)
}

func TestFilter(t *testing.T) {

	testCases := map[string]struct {
		input    string
		expected string
		filter   Filter
	}{
		"simple scalar": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: $(VAR)`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 5`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"VAR": int64(5),
				}),
				FieldSpec: types.FieldSpec{Path: "spec/replicas"},
			},
		},
		"non-string scalar": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 1`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 1`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"VAR": int64(5),
				}),
				FieldSpec: types.FieldSpec{Path: "spec/replicas"},
			},
		},
		"wrong path": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 1`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
spec:
  replicas: 1`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"VAR": int64(5),
				}),
				FieldSpec: types.FieldSpec{Path: "a/b/c"},
			},
		},
		"sequence": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
- $(FOO)
- $(BAR)
- $(BAZ)
- $(FOO)+$(BAR)
- $(BOOL)
- $(FLOAT)`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
- foo
- bar
- $(BAZ)
- foo+bar
- false
- 1.23`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"FOO":   "foo",
					"BAR":   "bar",
					"BOOL":  false,
					"FLOAT": 1.23,
				}),
				FieldSpec: types.FieldSpec{Path: "data"},
			},
		},
		"maps": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  FOO: $(FOO)
  BAR: $(BAR)
  BAZ: $(BAZ)
  PLUS: $(FOO)+$(BAR)`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  FOO: foo
  BAR: bar
  BAZ: $(BAZ)
  PLUS: foo+bar`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"FOO": "foo",
					"BAR": "bar",
				}),
				FieldSpec: types.FieldSpec{Path: "data"},
			},
		},
		"complicated case": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  slice1:
  - $(FOO)
  slice2:
    FOO: $(FOO)
    BAR: $(BAR)
    BOOL: false
    INT: 0
    SLICE:
    - $(FOO)`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  slice1:
  - $(FOO)
  slice2:
    FOO: foo
    BAR: bar
    BOOL: false
    INT: 0
    SLICE:
    - $(FOO)`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"FOO": "foo",
					"BAR": "bar",
				}),
				FieldSpec: types.FieldSpec{Path: "data/slice2"},
			},
		},
		"null value": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  FOO: null`,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  FOO: null`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					// no replacements!
				}),
				FieldSpec: types.FieldSpec{Path: "data/FOO"},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			if !assert.Equal(t,
				strings.TrimSpace(tc.expected),
				strings.TrimSpace(
					filtertest_test.RunFilter(t, tc.input, tc.filter))) {
				t.FailNow()
			}
		})
	}
}

func TestFilterUnhappy(t *testing.T) {
	testCases := map[string]struct {
		input         string
		expectedError string
		filter        Filter
	}{
		"non-string in sequence": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  slice:
  - false`,
			expectedError: `considering field 'data/slice' of object
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
  annotations:
    config.kubernetes.io/index: '0'
data:
  slice:
  - false
: invalid value type expect a string`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"VAR": int64(5),
				}),
				FieldSpec: types.FieldSpec{Path: "data/slice"},
			},
		},
		"invalid key in map": {
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
data:
  1: str`,
			expectedError: `considering field 'data' of object
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep
  annotations:
    config.kubernetes.io/index: '0'
data:
  1: str
: invalid map key: value='1', tag='` + yaml.NodeTagInt + `'`,
			filter: Filter{
				MappingFunc: makeMf(map[string]interface{}{
					"VAR": int64(5),
				}),
				FieldSpec: types.FieldSpec{Path: "data"},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			_, err := filtertest_test.RunFilterE(t, tc.input, tc.filter)
			if !assert.EqualError(t, err, tc.expectedError) {
				t.FailNow()
			}
		})
	}
}
