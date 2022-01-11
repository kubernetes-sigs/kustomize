// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/resource"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestOriginAppend(t *testing.T) {
	tests := []struct {
		in       *Origin
		path     string
		expected string
	}{
		{
			in: &Origin{
				Path: "prod",
			},
			path: "service.yaml",
			expected: `path: prod/service.yaml
`,
		},
		{
			in: &Origin{
				Path: "overlay/prod",
			},
			path: "github.com/kubernetes-sigs/kustomize/examples/multibases/dev/",
			expected: `path: examples/multibases/dev
repo: https://github.com/kubernetes-sigs/kustomize
`,
		},
	}
	for _, test := range tests {
		actual, err := test.in.Append(test.path).String()
		assert.NoError(t, err)
		assert.Equal(t, actual, test.expected)
	}
}

func TestOriginString(t *testing.T) {
	tests := []struct {
		in       *Origin
		expected string
	}{
		{
			in: &Origin{
				Path: "prod/service.yaml",
				Repo: "github.com/kubernetes-sigs/kustomize/examples/multibases/dev/",
				Ref:  "v1.0.6",
			},
			expected: `path: prod/service.yaml
repo: github.com/kubernetes-sigs/kustomize/examples/multibases/dev/
ref: v1.0.6
`,
		},
		{
			in: &Origin{
				Path: "prod/service.yaml",
				Repo: "github.com/kubernetes-sigs/kustomize/examples/multibases/dev/",
			},
			expected: `path: prod/service.yaml
repo: github.com/kubernetes-sigs/kustomize/examples/multibases/dev/
`,
		},
		{
			in: &Origin{
				Path: "prod/service.yaml",
			},
			expected: `path: prod/service.yaml
`,
		},
	}

	for _, test := range tests {
		actual, err := test.in.String()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}
}

func TestTransformationsString(t *testing.T) {
	origin1 := &Origin{
		Repo:         "github.com/myrepo",
		Ref:          "master",
		ConfiguredIn: "config.yaml",
		ConfiguredBy: kyaml.ResourceIdentifier{
			TypeMeta: kyaml.TypeMeta{
				APIVersion: "builtin",
				Kind:       "Generator",
			},
			NameMeta: kyaml.NameMeta{
				Name:      "my-name",
				Namespace: "my-namespace",
			},
		},
	}
	origin2 := &Origin{
		ConfiguredIn: "../base/config.yaml",
		ConfiguredBy: kyaml.ResourceIdentifier{
			TypeMeta: kyaml.TypeMeta{
				APIVersion: "builtin",
				Kind:       "Generator",
			},
			NameMeta: kyaml.NameMeta{
				Name:      "my-name",
				Namespace: "my-namespace",
			},
		},
	}
	tests := []struct {
		in       Transformations
		expected string
	}{
		{
			in: Transformations{origin1},
			expected: `- repo: github.com/myrepo
  ref: master
  configuredIn: config.yaml
  configuredBy:
    apiVersion: builtin
    kind: Generator
    name: my-name
    namespace: my-namespace
`,
		},
		{
			in: Transformations{origin1, origin2},
			expected: `- repo: github.com/myrepo
  ref: master
  configuredIn: config.yaml
  configuredBy:
    apiVersion: builtin
    kind: Generator
    name: my-name
    namespace: my-namespace
- configuredIn: ../base/config.yaml
  configuredBy:
    apiVersion: builtin
    kind: Generator
    name: my-name
    namespace: my-namespace
`,
		},
		{
			in: Transformations{},
			expected: `[]
`,
		},
	}
	for _, test := range tests {
		actual, err := test.in.String()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}
}
