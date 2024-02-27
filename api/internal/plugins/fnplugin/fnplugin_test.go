// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fnplugin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Test_GetFunctionSpec(t *testing.T) {
	factory := provider.NewDefaultDepProvider().GetResourceFactory()
	var tests = []struct {
		name       string
		resource   string
		catalogs   []framework.Catalog
		expectedFn string
		missingFn  bool
	}{
		// fn annotation, no catalogs
		{
			name: "fn from annotation",
			resource: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: test
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
`,
			expectedFn: `
container:
  image: foo:v1.0.0
		`,
			catalogs: []framework.Catalog{},
		},
		// from catalog, no annotation
		{
			name: "fn from catalog",
			resource: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: test
`,
			catalogs: []framework.Catalog{
				{
					Spec: framework.CatalogSpec{
						KrmFunctions: []framework.KrmFunctionDefinitionSpec{
							{
								Group: "example.com",
								Names: framework.KRMFunctionNames{
									Kind: "Example",
								},
								Versions: []framework.KRMFunctionVersion{
									{
										Name: "v1",
										Runtime: runtimeutil.FunctionSpec{
											Container: runtimeutil.ContainerSpec{
												Image: "foo:v1.0.0",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedFn: `
container:
  image: foo:v1.0.0
		`,
		},
		{
			// nil fn spec
			resource: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: test
`,
			catalogs:  []framework.Catalog{},
			missingFn: true,
		},
		// when both are provided, fn spec is taken from catalog.
		{
			name: "accept catalog fn spec when both are provided",
			resource: `
apiVersion: example.com/v1
kind: Example
metadata:
  name: test
  annotations:
    config.kubernetes.io/function: |-
      container:
        image: foo:v1.0.0
`,
			catalogs: []framework.Catalog{
				{
					Spec: framework.CatalogSpec{
						KrmFunctions: []framework.KrmFunctionDefinitionSpec{
							{
								Group: "example.com",
								Names: framework.KRMFunctionNames{
									Kind: "Example",
								},
								Versions: []framework.KRMFunctionVersion{
									{
										Name: "v1",
										Runtime: runtimeutil.FunctionSpec{
											Container: runtimeutil.ContainerSpec{
												Image: "foocatalog:v1.0.0",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedFn: `
container:
  image: foocatalog:v1.0.0
`,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			resource, err := factory.FromBytes([]byte(tt.resource))
			assert.NoError(t, err)
			fn, err := GetFunctionSpec(resource, tt.catalogs)
			if tt.missingFn {
				if err == nil && !assert.Nil(t, fn) {
					t.FailNow()
				}
				return
			}

			b, err := yaml.Marshal(fn)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(tt.expectedFn),
				strings.TrimSpace(string(b))) {
				t.FailNow()
			}
		})
	}
}
