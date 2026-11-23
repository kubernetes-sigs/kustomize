// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// listRecorder records the list kind and merge key the walker passes for each list.
type listRecorder struct {
	kinds     []ListKind
	mergeKeys []string
}

func (v *listRecorder) VisitMap(s Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	return s.Dest(), nil
}

func (v *listRecorder) VisitScalar(s Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	return s.Dest(), nil
}

func (v *listRecorder) VisitList(s Sources, rs *openapi.ResourceSchema, kind ListKind) (*yaml.RNode, error) {
	var key string
	if rs != nil {
		_, key = rs.PatchStrategyAndKey()
	}
	v.kinds = append(v.kinds, kind)
	v.mergeKeys = append(v.mergeKeys, key)
	return s.Dest(), nil
}

// The walker takes a field's schema from its line or head comment when present, and
// otherwise from OpenAPI; comment-free fields of unknown types get no schema.
func TestWalkerFieldSchema(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		kind     ListKind
		mergeKey string
	}{
		{
			name: "schema from field line comment",
			input: `apiVersion: example.com/v1
kind: Custom
containers: # {"items":{"$ref": "#/definitions/io.k8s.api.core.v1.Container"},"type":"array","x-kubernetes-patch-merge-key":"name","x-kubernetes-patch-strategy": "merge"}
- name: a
`,
			kind:     AssociativeList,
			mergeKey: "name",
		},
		{
			name: "no schema without field comment",
			input: `apiVersion: example.com/v1
kind: Custom
containers:
- name: a
`,
			kind: NonAssociateList,
		},
		{
			name: "schema from openapi for comment-free field",
			input: `apiVersion: v1
kind: Pod
spec:
  containers:
  - name: a
`,
			kind:     AssociativeList,
			mergeKey: "name",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := yaml.Parse(tc.input)
			require.NoError(t, err)
			v := &listRecorder{}
			_, err = Walker{Sources: []*yaml.RNode{node}, Visitor: v}.Walk()
			require.NoError(t, err)
			require.Len(t, v.kinds, 1)
			assert.Equal(t, tc.kind, v.kinds[0])
			assert.Equal(t, tc.mergeKey, v.mergeKeys[0])
		})
	}
}
