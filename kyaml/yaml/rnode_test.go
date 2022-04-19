// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/internal/forked/github.com/go-yaml/yaml"
)

func TestRNodeHasNilEntryInList(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		hasNil bool
		path   string
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp:  resultExpected{},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp:  resultExpected{},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{},
		},
		"listWithNil": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					nil,
				},
			},
			rsExp: resultExpected{
				hasNil: false, // TODO: This should be true.
				path:   "this/should/be/non-empty",
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			hasNil, path := rn.HasNilEntryInList()
			if tc.rsExp.hasNil {
				if !assert.True(t, hasNil) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.path, path) {
					t.FailNow()
				}
			} else {
				if !assert.False(t, hasNil) {
					t.FailNow()
				}
				if !assert.Empty(t, path) {
					t.FailNow()
				}
			}
		})
	}
}

func TestRNodeNewStringRNodeText(t *testing.T) {
	rn := NewStringRNode("cat")
	if !assert.Equal(t, `cat
`,
		rn.MustString()) {
		t.FailNow()
	}
}

func TestRNodeNewStringRNodeBinary(t *testing.T) {
	rn := NewStringRNode(string([]byte{
		0xff, // non-utf8
		0x68, // h
		0x65, // e
		0x6c, // l
		0x6c, // l
		0x6f, // o
	}))
	if !assert.Equal(t, `!!binary /2hlbGxv
`,
		rn.MustString()) {
		t.FailNow()
	}
}

func TestRNodeGetDataMap(t *testing.T) {
	emptyMap := map[string]string{}
	testCases := map[string]struct {
		theMap   map[string]interface{}
		expected map[string]string
	}{
		"actuallyNil": {
			theMap:   nil,
			expected: emptyMap,
		},
		"empty": {
			theMap:   map[string]interface{}{},
			expected: emptyMap,
		},
		"mostlyEmpty": {
			theMap: map[string]interface{}{
				"hey": "there",
			},
			expected: emptyMap,
		},
		"noNameConfigMap": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
			expected: emptyMap,
		},
		"configmap": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "winnie",
				},
				"data": map[string]string{
					"wine":   "cabernet",
					"truck":  "ford",
					"rocket": "falcon9",
					"planet": "mars",
					"city":   "brownsville",
				},
			},
			// order irrelevant, because assert.Equals is smart about maps.
			expected: map[string]string{
				"city":   "brownsville",
				"wine":   "cabernet",
				"planet": "mars",
				"rocket": "falcon9",
				"truck":  "ford",
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			m := rn.GetDataMap()
			if !assert.Equal(t, tc.expected, m) {
				t.FailNow()
			}
		})
	}
}

func TestRNodeGetValidatedDataMap(t *testing.T) {
	emptyMap := map[string]string{}
	testCases := map[string]struct {
		theMap         map[string]interface{}
		theAllowedKeys []string
		expected       map[string]string
		expectedError  error
	}{
		"nilResultEmptyKeys": {
			theMap:         nil,
			theAllowedKeys: []string{},
			expected:       emptyMap,
			expectedError:  nil,
		},
		"empty": {
			theMap:         map[string]interface{}{},
			theAllowedKeys: []string{},
			expected:       emptyMap,
			expectedError:  nil,
		},
		"expectedKeysMatch": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "winnie",
				},
				"data": map[string]string{
					"wine":   "cabernet",
					"truck":  "ford",
					"rocket": "falcon9",
					"planet": "mars",
					"city":   "brownsville",
				},
			},
			theAllowedKeys: []string{
				"wine",
				"truck",
				"rocket",
				"planet",
				"city",
				"plane",
				"country",
			},
			// order irrelevant, because assert.Equals is smart about maps.
			expected: map[string]string{
				"city":   "brownsville",
				"wine":   "cabernet",
				"planet": "mars",
				"rocket": "falcon9",
				"truck":  "ford",
			},
			expectedError: nil,
		},
		"unexpectedKeyInConfigMap": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "winnie",
				},
				"data": map[string]string{
					"wine":   "cabernet",
					"truck":  "ford",
					"rocket": "falcon9",
				},
			},
			theAllowedKeys: []string{
				"wine",
				"truck",
			},
			// order irrelevant, because assert.Equals is smart about maps.
			expected: map[string]string{
				"wine":   "cabernet",
				"rocket": "falcon9",
				"truck":  "ford",
			},
			expectedError: fmt.Errorf("an unexpected key (rocket) was found"),
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			m, err := rn.GetValidatedDataMap(tc.theAllowedKeys)
			if !assert.Equal(t, tc.expected, m) {
				t.FailNow()
			}
			if !assert.Equal(t, tc.expectedError, err) {
				t.FailNow()
			}
		})
	}
}

func TestRNodeSetDataMap(t *testing.T) {
	testCases := map[string]struct {
		theMap   map[string]interface{}
		input    map[string]string
		expected map[string]string
	}{
		"empty": {
			theMap: map[string]interface{}{},
			input: map[string]string{
				"wine":  "cabernet",
				"truck": "ford",
			},
			expected: map[string]string{
				"wine":  "cabernet",
				"truck": "ford",
			},
		},
		"replace": {
			theMap: map[string]interface{}{
				"foo": 3,
				"data": map[string]string{
					"rocket": "falcon9",
					"planet": "mars",
				},
			},
			input: map[string]string{
				"wine":  "cabernet",
				"truck": "ford",
			},
			expected: map[string]string{
				"wine":  "cabernet",
				"truck": "ford",
			},
		},
		"clear1": {
			theMap: map[string]interface{}{
				"foo": 3,
				"data": map[string]string{
					"rocket": "falcon9",
					"planet": "mars",
				},
			},
			input:    map[string]string{},
			expected: map[string]string{},
		},
		"clear2": {
			theMap: map[string]interface{}{
				"foo": 3,
				"data": map[string]string{
					"rocket": "falcon9",
					"planet": "mars",
				},
			},
			input:    nil,
			expected: map[string]string{},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			rn.SetDataMap(tc.input)
			m := rn.GetDataMap()
			if !assert.Equal(t, tc.expected, m) {
				t.FailNow()
			}
		})
	}
}

func TestRNodeGetValidatedMetadata(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		out    ResourceMeta
		errMsg string
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"mostlyEmpty": {
			theMap: map[string]interface{}{
				"hey": "there",
			},
			rsExp: resultExpected{
				errMsg: "missing Resource metadata",
			},
		},
		"noNameConfigMap": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
			},
			rsExp: resultExpected{
				errMsg: "missing metadata.name",
			},
		},
		"configmap": {
			theMap: testConfigMap,
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: ObjectMeta{
						NameMeta: NameMeta{
							Name: "winnie",
						},
					},
				},
			},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "List",
					},
				},
			},
		},
		"configmaplist": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMapList",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: ResourceMeta{
					TypeMeta: TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMapList",
					},
				},
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			m, err := rn.GetValidatedMetadata()
			if tc.rsExp.errMsg == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.out, m) {
					t.FailNow()
				}
			} else {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), tc.rsExp.errMsg) {
					t.FailNow()
				}
			}
		})
	}
}

func TestRNodeMapEmpty(t *testing.T) {
	newRNodeMap, err := NewRNode(nil).Map()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(newRNodeMap))
}

func TestRNodeMap(t *testing.T) {
	wn := NewRNode(nil)
	if err := wn.UnmarshalJSON([]byte(`{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "homer",
    "namespace": "simpsons"
  }
}`)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}

	expected := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "homer",
			"namespace": "simpsons",
		},
	}

	actual, err := wn.Map()
	assert.NoError(t, err)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
}

func TestRNodeFromMap(t *testing.T) {
	testConfigMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "winnie",
		},
	}
	type resultExpected struct {
		out string
		err error
	}

	testCases := map[string]struct {
		theMap map[string]interface{}
		rsExp  resultExpected
	}{
		"actuallyNil": {
			theMap: nil,
			rsExp: resultExpected{
				out: `{}`,
				err: nil,
			},
		},
		"empty": {
			theMap: map[string]interface{}{},
			rsExp: resultExpected{
				out: `{}`,
				err: nil,
			},
		},
		"mostlyEmpty": {
			theMap: map[string]interface{}{
				"hey": "there",
			},
			rsExp: resultExpected{
				out: `hey: there`,
				err: nil,
			},
		},
		"configmap": {
			theMap: testConfigMap,
			rsExp: resultExpected{
				out: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: winnie
`,
				err: nil,
			},
		},
		"list": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: `
apiVersion: v1
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
kind: List
`,
				err: nil,
			},
		},
		"configmaplist": {
			theMap: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMapList",
				"items": []interface{}{
					testConfigMap,
					testConfigMap,
				},
			},
			rsExp: resultExpected{
				out: `
apiVersion: v1
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: winnie
kind: ConfigMapList
`,
				err: nil,
			},
		},
	}

	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			rn, err := FromMap(tc.theMap)
			if tc.rsExp.err == nil {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tc.rsExp.out),
					strings.TrimSpace(rn.MustString())) {
					t.FailNow()
				}
			} else {
				if !assert.Error(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.rsExp.err, err) {
					t.FailNow()
				}
			}
		})
	}
}

// Test that non-UTF8 characters in comments don't cause failures
func TestRNode_GetMeta_UTF16(t *testing.T) {
	sr, err := Parse(`apiVersion: rbac.istio.io/v1alpha1
kind: ServiceRole
metadata:
  name: wildcard
  namespace: default
  # If set to [“*”], it refers to all services in the namespace
  annotations:
    foo: bar
spec:
  rules:
    # There is one service in default namespace, should not result in a validation error
    # If set to [“*”], it refers to all services in the namespace
    - services: ["*"]
      methods: ["GET", "HEAD"]
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := sr.GetMeta()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := ResourceMeta{
		TypeMeta: TypeMeta{
			APIVersion: "rbac.istio.io/v1alpha1",
			Kind:       "ServiceRole",
		},
		ObjectMeta: ObjectMeta{
			NameMeta: NameMeta{
				Name:      "wildcard",
				Namespace: "default",
			},
			Annotations: map[string]string{"foo": "bar"},
		},
	}
	if !assert.Equal(t, expected, actual) {
		t.FailNow()
	}
}

func TestDeAnchor(t *testing.T) {
	rn, err := Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: wildcard
data:
  color: &color-used blue
  feeling: *color-used
`)
	assert.NoError(t, err)
	assert.NoError(t, rn.DeAnchor())
	actual, err := rn.String()
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: wildcard
data:
  color: blue
  feeling: blue
`), strings.TrimSpace(actual))
}

func TestDeAnchorMerge(t *testing.T) {
	testCases := []struct {
		description string
		input       string
		expected    string
		expectedErr error
	}{
		// *********
		// Test Case
		// *********
		{
			description: "simplest merge tag",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    foo: bar
  primaryColor:
    <<: *color-used
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    foo: bar
  primaryColor:
    foo: bar
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "keep duplicated keys",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: "#FF0000"
  color: "#FF00FF"
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: "#FF0000"
  color: "#FF00FF"
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "keep json",
			input:       `{"apiVersion": "v1", "kind": "MergeTagTest", "spec": {"color": {"rgb": "#FF0000"}}}`,
			expected:    `{"apiVersion": "v1", "kind": "MergeTagTest", "spec": {"color": {"rgb": "#FF0000"}}}`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "keep comments",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    foo: bar
  primaryColor:
    # use same color because is pretty
    rgb: "#FF0000"
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    foo: bar
  primaryColor:
    # use same color because is pretty
    rgb: "#FF0000"
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "works with explicit merge tag",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    foo: bar
  primaryColor:
    !!merge <<: *color-used
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    foo: bar
  primaryColor:
    foo: bar
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "works with explicit long merge tag",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    foo: bar
  primaryColor:
    !<tag:yaml.org,2002:merge> "<<" : *color-used
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    foo: bar
  primaryColor:
    foo: bar
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "merging properties",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    rgb: "#FF0000"
  primaryColor:
    <<: *color-used
    pretty: true
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    rgb: "#FF0000"
  primaryColor:
    pretty: true
    rgb: "#FF0000"
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "overriding value",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color-used
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    <<: *color-used
    pretty: true
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    pretty: true
    rgb: "#FF0000"
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "returns error when defining multiple merge keys",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    rgb: "#FF0000"
    pretty: false
  primaryColor: &primary
    rgb: "#0000FF"
    alpha: 50
  secondaryColor:
    <<: *color
    <<: *primary
    secondary: true
`,
			expectedErr: fmt.Errorf("duplicate merge key"),
		},
		// *********
		// Test Case
		// *********
		{
			description: "merging multiple anchors with sequence node",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    rgb: "#FF0000"
    pretty: false
  primaryColor: &primary
    rgb: "#0000FF"
    alpha: 50
  secondaryColor:
    <<: [ *color, *primary ]
    secondary: true
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    rgb: "#0000FF"
    alpha: 50
  secondaryColor:
    secondary: true
    rgb: "#FF0000"
    alpha: 50
    pretty: false
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "merging inline map",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    <<: {"pretty": true}
    rgb: "#0000FF"
    alpha: 50
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    rgb: "#0000FF"
    alpha: 50
    "pretty": true
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "merging inline sequence map",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    <<: [ *color, {"name": "ugly blue"}]
    rgb: "#0000FF"
    alpha: 50
`,
			expected: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color:
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    rgb: "#0000FF"
    alpha: 50
    "name": "ugly blue"
    pretty: false
`,
		},
		// *********
		// Test Case
		// *********
		{
			description: "error on nested lists on merges",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    rgb: "#FF0000"
    pretty: false
  primaryColor:
    <<: [ *color, [{"name": "ugly blue"}]]
    rgb: "#0000FF"
    alpha: 50
`,
			expectedErr: fmt.Errorf("invalid map merge: received a nested sequence"),
		},
		// *********
		// Test Case
		// *********
		{
			description: "error on non-map references on merges",
			input: `
apiVersion: v1
kind: MergeTagTest
metadata:
  name: test
data:
  color: &color
    - rgb: "#FF0000"
      pretty: false
  primaryColor:
    <<: [ *color, [{"name": "ugly blue"}]]
    rgb: "#0000FF"
    alpha: 50
`,
			expectedErr: fmt.Errorf("invalid map merge: received alias for a non-map value"),
		},
		// *********
		// Test Case
		// *********
		{
			description: "merging on a list",
			input: `
apiVersion: v1
kind: MergeTagTestList
items:
- apiVersion: v1
  kind: MergeTagTest
  metadata:
    name: test
  spec: &merge-spec
    something: true
- apiVersion: v1
  kind: MergeTagTest
  metadata:
    name: test
  spec:
    <<: *merge-spec
`,
			expected: `
apiVersion: v1
kind: MergeTagTestList
items:
- apiVersion: v1
  kind: MergeTagTest
  metadata:
    name: test
  spec:
    something: true
- apiVersion: v1
  kind: MergeTagTest
  metadata:
    name: test
  spec:
    something: true
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			rn, err := Parse(tc.input)

			assert.NoError(t, err)
			err = rn.DeAnchor()
			if tc.expectedErr == nil {
				assert.NoError(t, err)
				actual, err := rn.String()
				assert.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual))
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}
		})
	}
}
func TestRNode_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		output   string
	}{
		{
			testName: "simple document",
			input:    `{"hello":"world"}`,
			output:   `hello: world`,
		},
		{
			testName: "nested structure",
			input: `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "my-deployment",
    "namespace": "default"
  }
}
`,
			output: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: default
`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			instance := &RNode{}
			err := instance.UnmarshalJSON([]byte(tc.input))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.String()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tc.output), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}

func TestRNode_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		ydoc string
		want string
	}{
		{
			name: "object",
			ydoc: `
hello: world
`,
			want: `{"hello":"world"}`,
		},
		{
			name: "array",
			ydoc: `
- name: s1
- name: s2
`,
			want: `[{"name":"s1"},{"name":"s2"}]`,
		},
	}
	for idx := range tests {
		tt := tests[idx]
		t.Run(tt.name, func(t *testing.T) {
			instance, err := Parse(tt.ydoc)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			actual, err := instance.MarshalJSON()
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if !assert.Equal(t,
				strings.TrimSpace(tt.want), strings.TrimSpace(string(actual))) {
				t.FailNow()
			}
		})
	}
}

func TestConvertJSONToYamlNode(t *testing.T) {
	inputJSON := `{"type": "string", "maxLength": 15, "enum": ["allowedValue1", "allowedValue2"]}`
	expected := `enum:
- allowedValue1
- allowedValue2
maxLength: 15
type: string
`

	node, err := ConvertJSONToYamlNode(inputJSON)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := node.String()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, expected, actual)
}

func TestIsMissingOrNull(t *testing.T) {
	if !IsMissingOrNull(nil) {
		t.Fatalf("input: nil")
	}
	// missing value or null value
	if !IsMissingOrNull(NewRNode(nil)) {
		t.Fatalf("input: nil value")
	}

	if IsMissingOrNull(NewScalarRNode("foo")) {
		t.Fatalf("input: valid node")
	}
	// node with NullNodeTag
	if !IsMissingOrNull(MakeNullNode()) {
		t.Fatalf("input: with NullNodeTag")
	}

	// empty array. empty array is not expected as empty
	if IsMissingOrNull(NewListRNode()) {
		t.Fatalf("input: empty array")
	}

	// array with 1 item
	node := NewListRNode("foo")
	if IsMissingOrNull(node) {
		t.Fatalf("input: array with 1 item")
	}

	// delete the item in array
	node.value.Content = nil
	if IsMissingOrNull(node) {
		t.Fatalf("input: empty array")
	}
}

func TestCopy(t *testing.T) {
	rn := RNode{
		fieldPath: []string{"fp1", "fp2"},
		value: &Node{
			Kind: 200,
		},
		Match: []string{"m1", "m2"},
	}
	rnC := rn.Copy()
	if !reflect.DeepEqual(&rn, rnC) {
		t.Fatalf("copy %v is not deep equal to %v", rnC, rn)
	}
	tmp := rn.value.Kind
	rn.value.Kind = 666
	if reflect.DeepEqual(rn, rnC) {
		t.Fatalf("changing component should break equality")
	}
	rn.value.Kind = tmp
	if !reflect.DeepEqual(&rn, rnC) {
		t.Fatalf("should be back to normal")
	}
	rn.fieldPath[0] = "Different"
	if reflect.DeepEqual(rn, rnC) {
		t.Fatalf("changing fieldpath should break equality")
	}
}

func TestFieldRNodes(t *testing.T) {
	testCases := []struct {
		testName string
		input    string
		output   []string
		err      string
	}{
		{
			testName: "simple document",
			input: `apiVersion: example.com/v1beta1
kind: Example1
spec:
  list:
  - "a"
  - "b"
  - "c"`,
			output: []string{"apiVersion", "kind", "spec", "list"},
		},
		{
			testName: "non mapping node error",
			input:    `apiVersion`,
			err:      "wrong Node Kind for  expected: MappingNode was ScalarNode: value: {apiVersion}",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			rNode, err := Parse(tc.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			fieldRNodes, err := rNode.FieldRNodes()
			if tc.err == "" {
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else {
				if !assert.Equal(t, tc.err, err.Error()) {
					t.FailNow()
				}
			}
			for i := range fieldRNodes {
				actual, err := fieldRNodes[i].String()
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t, tc.output[i], strings.TrimSpace(actual)) {
					t.FailNow()
				}
			}
		})
	}
}

func TestIsEmptyMap(t *testing.T) {
	node := NewMapRNode(nil)
	// empty map
	if !IsEmptyMap(node) {
		t.Fatalf("input: empty map")
	}
	// map with 1 item
	node = NewMapRNode(&map[string]string{
		"foo": "bar",
	})
	if IsEmptyMap(node) {
		t.Fatalf("input: map with 1 item")
	}
	// delete the item in map
	node.value.Content = nil
	if !IsEmptyMap(node) {
		t.Fatalf("input: empty map")
	}
}

func TestIsNil(t *testing.T) {
	var rn *RNode

	if !rn.IsNil() {
		t.Fatalf("uninitialized RNode should be nil")
	}

	if !NewRNode(nil).IsNil() {
		t.Fatalf("missing value YNode should be nil")
	}

	if MakeNullNode().IsNil() {
		t.Fatalf("value tagged null is not nil")
	}

	if NewMapRNode(nil).IsNil() {
		t.Fatalf("empty map not nil")
	}

	if NewListRNode().IsNil() {
		t.Fatalf("empty list not nil")
	}
}

func TestIsTaggedNull(t *testing.T) {
	var rn *RNode

	if rn.IsTaggedNull() {
		t.Fatalf("nil RNode cannot be tagged")
	}

	if NewRNode(nil).IsTaggedNull() {
		t.Fatalf("bare RNode should not be tagged")
	}

	if !MakeNullNode().IsTaggedNull() {
		t.Fatalf("a null node is tagged null by definition")
	}

	if NewMapRNode(nil).IsTaggedNull() {
		t.Fatalf("empty map should not be tagged null")
	}

	if NewListRNode().IsTaggedNull() {
		t.Fatalf("empty list should not be tagged null")
	}
}

func TestRNodeIsNilOrEmpty(t *testing.T) {
	var rn *RNode

	if !rn.IsNilOrEmpty() {
		t.Fatalf("uninitialized RNode should be empty")
	}

	if !NewRNode(nil).IsNilOrEmpty() {
		t.Fatalf("missing value YNode should be empty")
	}

	if !MakeNullNode().IsNilOrEmpty() {
		t.Fatalf("value tagged null should be empty")
	}

	if !NewMapRNode(nil).IsNilOrEmpty() {
		t.Fatalf("empty map should be empty")
	}

	if NewMapRNode(&map[string]string{"foo": "bar"}).IsNilOrEmpty() {
		t.Fatalf("non-empty map should not be empty")
	}

	if !NewListRNode().IsNilOrEmpty() {
		t.Fatalf("empty list should be empty")
	}

	if NewListRNode("foo").IsNilOrEmpty() {
		t.Fatalf("non-empty list should not be empty")
	}

	if !NewRNode(&Node{}).IsNilOrEmpty() {
		t.Fatalf("zero YNode should be empty")
	}
}

const deploymentJSON = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "homer",
    "namespace": "simpsons",
    "labels": {
      "fruit": "apple",
      "veggie": "carrot"
    },
    "annotations": {
      "area": "51",
      "greeting": "Take me to your leader."
    }
  }
}
`

func TestRNodeSetNamespace(t *testing.T) {
	n := NewRNode(nil)
	assert.Equal(t, "", n.GetNamespace())
	assert.NoError(t, n.SetNamespace(""))
	assert.Equal(t, "", n.GetNamespace())
	if err := n.UnmarshalJSON([]byte(deploymentJSON)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	assert.NoError(t, n.SetNamespace("flanders"))
	assert.Equal(t, "flanders", n.GetNamespace())
}

func TestRNodeSetLabels(t *testing.T) {
	n := NewRNode(nil)
	assert.Equal(t, 0, len(n.GetLabels()))
	assert.NoError(t, n.SetLabels(map[string]string{}))
	assert.Equal(t, 0, len(n.GetLabels()))

	n = NewRNode(nil)
	if err := n.UnmarshalJSON([]byte(deploymentJSON)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	labels := n.GetLabels()
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, "apple", labels["fruit"])
	assert.Equal(t, "carrot", labels["veggie"])
	assert.NoError(t, n.SetLabels(map[string]string{
		"label1": "foo",
		"label2": "bar",
	}))
	labels = n.GetLabels()
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, "foo", labels["label1"])
	assert.Equal(t, "bar", labels["label2"])
	assert.NoError(t, n.SetLabels(map[string]string{}))
	assert.Equal(t, 0, len(n.GetLabels()))
}

func TestRNodeGetAnnotations(t *testing.T) {
	n := NewRNode(nil)
	assert.Equal(t, 0, len(n.GetAnnotations()))
	assert.NoError(t, n.SetAnnotations(map[string]string{}))
	assert.Equal(t, 0, len(n.GetAnnotations()))

	n = NewRNode(nil)
	if err := n.UnmarshalJSON([]byte(deploymentJSON)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	annotations := n.GetAnnotations()
	assert.Equal(t, 2, len(annotations))
	assert.Equal(t, "51", annotations["area"])
	assert.Equal(t, "Take me to your leader.", annotations["greeting"])
	assert.NoError(t, n.SetAnnotations(map[string]string{
		"annotation1": "foo",
		"annotation2": "bar",
	}))
	annotations = n.GetAnnotations()
	assert.Equal(t, 2, len(annotations))
	assert.Equal(t, "foo", annotations["annotation1"])
	assert.Equal(t, "bar", annotations["annotation2"])
	assert.NoError(t, n.SetAnnotations(map[string]string{}))
	assert.Equal(t, 0, len(n.GetAnnotations()))
}

func TestRNodeMatchesAnnotationSelector(t *testing.T) {
	rn := NewRNode(nil)
	assert.NoError(t, rn.UnmarshalJSON([]byte(deploymentJSON)))
	testcases := map[string]struct {
		selector string
		matches  bool
		errMsg   string
	}{
		"select_01": {
			selector: ".*",
			matches:  false,
			errMsg:   "name part must consist of alphanumeric character",
		},
		"select_02": {
			selector: "area=51",
			matches:  true,
		},
		"select_03": {
			selector: "area=florida",
			matches:  false,
		},
		"select_04": {
			selector: "area in (disneyland, 51, iowa)",
			matches:  true,
		},
		"select_05": {
			selector: "area in (disneyland, iowa)",
			matches:  false,
		},
		"select_06": {
			selector: "area notin (disneyland, 51, iowa)",
			matches:  false,
		},
	}
	for n, tc := range testcases {
		gotMatch, err := rn.MatchesAnnotationSelector(tc.selector)
		if tc.errMsg == "" {
			assert.NoError(t, err)
			assert.Equalf(
				t, tc.matches, gotMatch, "test=%s selector=%v", n, tc.selector)
		} else {
			assert.Truef(
				t, strings.Contains(err.Error(), tc.errMsg),
				"test=%s selector=%v", n, tc.selector)
		}
	}
}

func TestRNodeMatchesLabelSelector(t *testing.T) {
	rn := NewRNode(nil)
	assert.NoError(t, rn.UnmarshalJSON([]byte(deploymentJSON)))
	testcases := map[string]struct {
		selector string
		matches  bool
		errMsg   string
	}{
		"select_01": {
			selector: ".*",
			matches:  false,
			errMsg:   "name part must consist of alphanumeric character",
		},
		"select_02": {
			selector: "fruit=apple",
			matches:  true,
		},
		"select_03": {
			selector: "fruit=banana",
			matches:  false,
		},
		"select_04": {
			selector: "fruit in (orange, banana, apple)",
			matches:  true,
		},
		"select_05": {
			selector: "fruit in (orange, banana)",
			matches:  false,
		},
		"select_06": {
			selector: "fruit notin (orange, banana, apple)",
			matches:  false,
		},
	}
	for n, tc := range testcases {
		gotMatch, err := rn.MatchesLabelSelector(tc.selector)
		if tc.errMsg == "" {
			assert.NoError(t, err)
			assert.Equalf(
				t, tc.matches, gotMatch, "test=%s selector=%v", n, tc.selector)
		} else {
			assert.Truef(
				t, strings.Contains(err.Error(), tc.errMsg),
				"test=%s selector=%v", n, tc.selector)
		}
	}
}

const (
	deploymentLittleJson = `{"apiVersion":"apps/v1","kind":"Deployment",` +
		`"metadata":{"name":"homer","namespace":"simpsons"}}`

	deploymentBiggerJson = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "homer",
    "namespace": "simpsons",
    "labels": {
      "fruit": "apple",
      "veggie": "carrot"
    },
    "annotations": {
      "area": "51",
      "greeting": "Take me to your leader."
    }
  },
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "env": [
              {
                "name": "CM_FOO",
                "valueFrom": {
                  "configMapKeyRef": {
                    "key": "somekey",
                    "name": "myCm"
                  }
                }
              },
              {
                "name": "SECRET_FOO",
                "valueFrom": {
                  "secretKeyRef": {
                    "key": "someKey",
                    "name": "mySecret"
                  }
                }
              }
            ],
            "image": "nginx:1.7.9",
            "name": "nginx"
          }
        ]
      }
    }
  }
}
`
	bigMapYaml = `Kind: Service
complextree:
- field1:
  - boolfield: true
    floatsubfield: 1.01
    intsubfield: 1010
    stringsubfield: idx1010
  - boolfield: false
    floatsubfield: 1.011
    intsubfield: 1011
    stringsubfield: idx1011
  field2:
  - boolfield: true
    floatsubfield: 1.02
    intsubfield: 1020
    stringsubfield: idx1020
  - boolfield: false
    floatsubfield: 1.021
    intsubfield: 1021
    stringsubfield: idx1021
- field1:
  - boolfield: true
    floatsubfield: 1.11
    intsubfield: 1110
    stringsubfield: idx1110
  - boolfield: false
    floatsubfield: 1.111
    intsubfield: 1111
    stringsubfield: idx1111
  field2:
  - boolfield: true
    floatsubfield: 1.112
    intsubfield: 1120
    stringsubfield: idx1120
  - boolfield: false
    floatsubfield: 1.1121
    intsubfield: 1121
    stringsubfield: idx1121
metadata:
  labels:
    app: application-name
  name: service-name
spec:
  ports:
    port: 80
that:
- idx0
- idx1
- idx2
- idx3
these:
- field1:
  - idx010
  - idx011
  field2:
  - idx020
  - idx021
- field1:
  - idx110
  - idx111
  field2:
  - idx120
  - idx121
- field1:
  - idx210
  - idx211
  field2:
  - idx220
  - idx221
this:
  is:
    aBool: true
    aFloat: 1.001
    aNilValue: null
    aNumber: 1000
    anEmptyMap: {}
    anEmptySlice: []
those:
- field1: idx0foo
  field2: idx0bar
- field1: idx1foo
  field2: idx1bar
- field1: idx2foo
  field2: idx2bar
`
)

func TestGetFieldValueReturnsMap(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := map[string]interface{}{
		"fruit":  "apple",
		"veggie": "carrot",
	}
	actual, err := rn.GetFieldValue("metadata.labels")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
}

func TestGetFieldValueReturnsStuff(t *testing.T) {
	wn := NewRNode(nil)
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := []interface{}{
		map[string]interface{}{
			"env": []interface{}{
				map[string]interface{}{
					"name": "CM_FOO",
					"valueFrom": map[string]interface{}{
						"configMapKeyRef": map[string]interface{}{
							"key":  "somekey",
							"name": "myCm",
						},
					},
				},
				map[string]interface{}{
					"name": "SECRET_FOO",
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"key":  "someKey",
							"name": "mySecret",
						},
					},
				},
			},
			"image": string("nginx:1.7.9"),
			"name":  string("nginx"),
		},
	}
	actual, err := wn.GetFieldValue("spec.template.spec.containers")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
	// Cannot go deeper yet.
	_, err = wn.GetFieldValue("spec.template.spec.containers.env")
	if err == nil {
		t.Fatalf("expected err %v", err)
	}
}

func makeBigMap() map[string]interface{} {
	return map[string]interface{}{
		"Kind": "Service",
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"app": "application-name",
			},
			"name": "service-name",
		},
		"spec": map[string]interface{}{
			"ports": map[string]interface{}{
				"port": int64(80),
			},
		},
		"this": map[string]interface{}{
			"is": map[string]interface{}{
				"aNumber":      int64(1000),
				"aFloat":       float64(1.001),
				"aNilValue":    nil,
				"aBool":        true,
				"anEmptyMap":   map[string]interface{}{},
				"anEmptySlice": []interface{}{},
				/*
					TODO: test for unrecognizable (e.g. a function)
						"unrecognizable": testing.InternalExample{
							Name: "fooBar",
						},
				*/
			},
		},
		"that": []interface{}{
			"idx0",
			"idx1",
			"idx2",
			"idx3",
		},
		"those": []interface{}{
			map[string]interface{}{
				"field1": "idx0foo",
				"field2": "idx0bar",
			},
			map[string]interface{}{
				"field1": "idx1foo",
				"field2": "idx1bar",
			},
			map[string]interface{}{
				"field1": "idx2foo",
				"field2": "idx2bar",
			},
		},
		"these": []interface{}{
			map[string]interface{}{
				"field1": []interface{}{"idx010", "idx011"},
				"field2": []interface{}{"idx020", "idx021"},
			},
			map[string]interface{}{
				"field1": []interface{}{"idx110", "idx111"},
				"field2": []interface{}{"idx120", "idx121"},
			},
			map[string]interface{}{
				"field1": []interface{}{"idx210", "idx211"},
				"field2": []interface{}{"idx220", "idx221"},
			},
		},
		"complextree": []interface{}{
			map[string]interface{}{
				"field1": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1010",
						"intsubfield":    int64(1010),
						"floatsubfield":  float64(1.010),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1011",
						"intsubfield":    int64(1011),
						"floatsubfield":  float64(1.011),
						"boolfield":      false,
					},
				},
				"field2": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1020",
						"intsubfield":    int64(1020),
						"floatsubfield":  float64(1.020),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1021",
						"intsubfield":    int64(1021),
						"floatsubfield":  float64(1.021),
						"boolfield":      false,
					},
				},
			},
			map[string]interface{}{
				"field1": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1110",
						"intsubfield":    int64(1110),
						"floatsubfield":  float64(1.110),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1111",
						"intsubfield":    int64(1111),
						"floatsubfield":  float64(1.111),
						"boolfield":      false,
					},
				},
				"field2": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1120",
						"intsubfield":    int64(1120),
						"floatsubfield":  float64(1.1120),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1121",
						"intsubfield":    int64(1121),
						"floatsubfield":  float64(1.1121),
						"boolfield":      false,
					},
				},
			},
		},
	}
}
func TestBasicYamlOperationFromMap(t *testing.T) {
	bytes, err := Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	if string(bytes) != bigMapYaml {
		t.Fatalf("unexpected string equality")
	}
	rNode, err := Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rNodeString := rNode.MustString()
	// The result from MustString has more indentation
	// than bigMapYaml.
	rNodeStrings := strings.Split(rNodeString, "\n")
	bigMapStrings := strings.Split(bigMapYaml, "\n")
	if len(rNodeStrings) != len(bigMapStrings) {
		t.Fatalf("line count mismatch")
	}
	for i := range rNodeStrings {
		s1 := strings.TrimSpace(rNodeStrings[i])
		s2 := strings.TrimSpace(bigMapStrings[i])
		if s1 != s2 {
			t.Fatalf("expected '%s'=='%s'", s1, s2)
		}
	}
}

func TestGetFieldValueReturnsSlice(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rNode, err := Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	expected := []interface{}{"idx0", "idx1", "idx2", "idx3"}
	actual, err := rNode.GetFieldValue("that")
	if err != nil {
		t.Fatalf("error getting slice: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual slice does not deep equal expected slice:\n%v", diff)
	}
}

func TestGetFieldValueReturnsSliceOfMappings(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rn, err := Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	expected := []interface{}{
		map[string]interface{}{
			"field1": "idx0foo",
			"field2": "idx0bar",
		},
		map[string]interface{}{
			"field1": "idx1foo",
			"field2": "idx1bar",
		},
		map[string]interface{}{
			"field1": "idx2foo",
			"field2": "idx2bar",
		},
	}
	actual, err := rn.GetFieldValue("those")
	if err != nil {
		t.Fatalf("error getting slice: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual slice does not deep equal expected slice:\n%v", diff)
	}
}

func TestGetFieldValueReturnsString(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	actual, err := rn.GetFieldValue("metadata.labels.fruit")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	v, ok := actual.(string)
	if !ok || v != "apple" {
		t.Fatalf("unexpected value '%v'", actual)
	}
}

func TestGetFieldValueResolvesAlias(t *testing.T) {
	yamlWithAlias := `
foo: &a theValue
bar: *a
`
	rn, err := Parse(yamlWithAlias)
	if err != nil {
		t.Fatalf("unexpected yaml parse error: %v", err)
	}
	actual, err := rn.GetFieldValue("bar")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	v, ok := actual.(string)
	if !ok || v != "theValue" {
		t.Fatalf("unexpected value '%v'", actual)
	}
}

func TestGetString(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := "carrot"
	actual, err := rn.GetString("metadata.labels.veggie")
	if err != nil {
		t.Fatalf("error getting string: %v", err)
	}
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestGetSlice(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rn, err := Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	expected := []interface{}{"idx0", "idx1", "idx2", "idx3"}
	actual, err := rn.GetSlice("that")
	if err != nil {
		t.Fatalf("error getting slice: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual slice does not deep equal expected slice:\n%v", diff)
	}
}

func TestRoundTripJSON(t *testing.T) {
	rn := NewRNode(nil)
	err := rn.UnmarshalJSON([]byte(deploymentLittleJson))
	if err != nil {
		t.Fatalf("unexpected UnmarshalJSON err: %v", err)
	}
	data, err := rn.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected MarshalJSON err: %v", err)
	}
	if actual := string(data); actual != deploymentLittleJson {
		t.Fatalf("expected %s, got %s", deploymentLittleJson, actual)
	}
}

func TestGettingFields(t *testing.T) {
	rn := NewRNode(nil)
	err := rn.UnmarshalJSON([]byte(deploymentBiggerJson))
	if err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := "Deployment"
	actual := rn.GetKind()
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "homer"
	actual = rn.GetName()
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	actualMap := rn.GetLabels()
	v, ok := actualMap["fruit"]
	if !ok || v != "apple" {
		t.Fatalf("unexpected labels '%v'", actualMap)
	}
	actualMap = rn.GetAnnotations()
	v, ok = actualMap["greeting"]
	if !ok || v != "Take me to your leader." {
		t.Fatalf("unexpected annotations '%v'", actualMap)
	}
}

func TestMapEmpty(t *testing.T) {
	newNodeMap, err := NewRNode(nil).Map()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(newNodeMap))
}

func TestMap(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentLittleJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}

	expected := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "homer",
			"namespace": "simpsons",
		},
	}

	actual, err := rn.Map()
	assert.NoError(t, err)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
}

func TestSetName(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	err := rn.SetName("marge")
	require.NoError(t, err)
	if expected, actual := "marge", rn.GetName(); expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestSetNamespace(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	err := rn.SetNamespace("flanders")
	require.NoError(t, err)
	meta, err := rn.GetMeta()
	require.NoError(t, err)
	if expected, actual := "flanders", meta.Namespace; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestSetLabels(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	assert.NoError(t, rn.SetLabels(map[string]string{
		"label1": "foo",
		"label2": "bar",
	}))
	labels := rn.GetLabels()
	if expected, actual := 2, len(labels); expected != actual {
		t.Fatalf("expected '%d', got '%d'", expected, actual)
	}
	if expected, actual := "foo", labels["label1"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "bar", labels["label2"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestGetAnnotations(t *testing.T) {
	rn := NewRNode(nil)
	if err := rn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	assert.NoError(t, rn.SetAnnotations(map[string]string{
		"annotation1": "foo",
		"annotation2": "bar",
	}))
	annotations := rn.GetAnnotations()
	if expected, actual := 2, len(annotations); expected != actual {
		t.Fatalf("expected '%d', got '%d'", expected, actual)
	}
	if expected, actual := "foo", annotations["annotation1"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "bar", annotations["annotation2"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestGetFieldValueWithDot(t *testing.T) {
	//t.Skip()

	const input = `
kind: Pod
metadata:
  name: hello-world
  labels:
    app: hello-world-app
    foo.appname: hello-world-foo
`
	data, err := Parse(input)
	require.NoError(t, err)

	labelRNode, err := data.Pipe(Lookup("metadata", "labels"))
	require.NoError(t, err)

	app, err := labelRNode.GetFieldValue("app")
	require.NoError(t, err)
	require.Equal(t, "hello-world-app", app)

	// TODO: doesn't currently work; we expect to be able to escape the dot in future
	// https://github.com/kubernetes-sigs/kustomize/issues/4487
	fooAppName, err := labelRNode.GetFieldValue(`foo\.appname`)
	require.NoError(t, err)
	require.Equal(t, "hello-world-foo", fooAppName) // no field named 'foo.appname'
}
