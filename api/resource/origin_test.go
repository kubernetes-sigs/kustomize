// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/api/resource"
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
		actual := test.in.Append(test.path).String()
		if actual != test.expected {
			t.Fatalf("Expected %v, but got %v\n", test.expected, actual)
		}
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
		if test.in.String() != test.expected {
			t.Fatalf("Expected %v, but got %v\n", test.expected, test.in.String())
		}
	}
}
