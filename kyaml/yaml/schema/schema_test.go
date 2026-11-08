// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	. "sigs.k8s.io/kustomize/kyaml/yaml/schema"
)

func TestIsAssociativeNoSchema(t *testing.T) {
	assert.False(t, IsAssociative(nil, []*yaml.RNode{}, false))
}

func makeSchema() *spec.Schema {
	return &spec.Schema{
		VendorExtensible: spec.VendorExtensible{
			Extensions: make(map[string]interface{}),
		},
	}
}

func TestIsAssociativeSimpleStrategy(t *testing.T) {
	s := makeSchema()
	s.Extensions["x-kubernetes-patch-merge-key"] = "name"
	s.Extensions["x-kubernetes-patch-strategy"] = "merge"
	assert.True(
		t,
		IsAssociative(
			&openapi.ResourceSchema{Schema: s},
			[]*yaml.RNode{}, false))
}

func TestIsAssociativeMultipleStrategy(t *testing.T) {
	s := makeSchema()
	s.Extensions["x-kubernetes-patch-merge-key"] = "name"
	s.Extensions["x-kubernetes-patch-strategy"] = "retainKeys,merge"
	assert.True(
		t,
		IsAssociative(
			&openapi.ResourceSchema{Schema: s},
			[]*yaml.RNode{}, false))
}

func TestIsAssociativeCRDListTopology(t *testing.T) {
	testCases := []struct {
		name       string
		extensions map[string]interface{}
		want       bool
	}{
		{
			name: "map",
			extensions: map[string]interface{}{
				"x-kubernetes-list-type":     "map",
				"x-kubernetes-list-map-keys": []interface{}{"name"},
			},
			want: true,
		},
		{
			name: "atomic",
			extensions: map[string]interface{}{
				"x-kubernetes-list-type": "atomic",
			},
			want: false,
		},
		{
			name: "set",
			extensions: map[string]interface{}{
				"x-kubernetes-list-type": "set",
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := makeSchema()
			s.Extensions = tc.extensions
			assert.Equal(t, tc.want, IsAssociative(
				&openapi.ResourceSchema{Schema: s}, []*yaml.RNode{}, false))
		})
	}
}
