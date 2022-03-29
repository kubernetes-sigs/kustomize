// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

var demoFunctionDefinition = `
apiVersion: config.kubernetes.io/v1alpha1
kind: KRMFunctionDefinition
metadata:
  name: demos.example.io
spec:
  group: example.io
  names:
    kind: Demo
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            type: string
          color:
            type: string
          kind:
            type: string
          metadata:
            type: object
        required:
        - color
        type: object
  - name: v1alpha2
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            type: string
          flavor:
            type: string
          kind:
            type: string
          metadata:
            type: object
        required:
        - flavor
        type: object
`

var demoCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: demos.example.io
spec:
  group: example.io
  names:
    kind: Demo
    listKind: DemoList
    plural: demos
    singular: demo
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: 
			  https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          color:
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: 
			  https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
        required:
        - color
        type: object
    served: true
    storage: true
`

func TestSchemaFromFunctionDefinition(t *testing.T) {
	tests := []struct {
		name      string
		gvk       resid.Gvk
		data      string
		wantProps []string
		wantErr   string
	}{
		{
			name:      "demo KRMFunctionDefinition extract v1alpha1",
			gvk:       resid.NewGvk("example.io", "v1alpha1", "Demo"),
			data:      demoFunctionDefinition,
			wantProps: []string{"apiVersion", "kind", "metadata", "color"},
		}, {
			name:      "demo KRMFunctionDefinition extract v1alpha2",
			gvk:       resid.NewGvk("example.io", "v1alpha2", "Demo"),
			data:      demoFunctionDefinition,
			wantProps: []string{"apiVersion", "kind", "metadata", "flavor"},
		}, {
			name:      "works with CustomResourceDefinition",
			gvk:       resid.NewGvk("example.io", "v1alpha1", "Demo"),
			data:      demoCRD,
			wantProps: []string{"apiVersion", "kind", "metadata", "color"},
		}, {
			name:    "group mismatch",
			gvk:     resid.NewGvk("example.com", "v1alpha2", "Demo"),
			data:    demoFunctionDefinition,
			wantErr: "KRMFunctionDefinition does not define Demo.v1alpha2.example.com (defines: [Demo.v1alpha1.example.io Demo.v1alpha2.example.io])",
		}, {
			name:    "version mismatch",
			gvk:     resid.NewGvk("example.io", "v1alpha3", "Demo"),
			data:    demoFunctionDefinition,
			wantErr: "KRMFunctionDefinition does not define Demo.v1alpha3.example.io (defines: [Demo.v1alpha1.example.io Demo.v1alpha2.example.io])",
		}, {
			name:    "kind mismatch",
			gvk:     resid.NewGvk("example.io", "v1alpha2", "Demonstration"),
			data:    demoFunctionDefinition,
			wantErr: "KRMFunctionDefinition does not define Demonstration.v1alpha2.example.io (defines: [Demo.v1alpha1.example.io Demo.v1alpha2.example.io])",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SchemaFromFunctionDefinition(tt.gvk, tt.data)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				var gotProps []string
				for prop := range got.Properties {
					gotProps = append(gotProps, prop)
				}
				sort.Strings(tt.wantProps)
				sort.Strings(gotProps)
				assert.Equal(t, gotProps, tt.wantProps)
			}
		})
	}
}
