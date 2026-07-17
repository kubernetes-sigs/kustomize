// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package builtinopenapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func TestBundleValidate(t *testing.T) {
	valid := func() Bundle {
		return Bundle{
			FormatVersion:   FormatVersion,
			Coverage:        Coverage{Floor: "v1.21.2", Ceiling: "v1.21.2"},
			SelectionPolicy: SelectionPolicy,
			Sources: []Source{{
				KubernetesVersion: "v1.21.2",
				SHA256:            "5d171b55e9601912807a870d73ffe70bb306f5889a00e76986042a0f2d7b6bc2",
			}},
			Definitions: spec.Definitions{"definition": {}},
			Resources: []Resource{{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Definition: "definition",
				Scope:      ScopeNamespaced,
			}},
		}
	}

	tests := map[string]func(*Bundle){
		"format":   func(bundle *Bundle) { bundle.FormatVersion++ },
		"coverage": func(bundle *Bundle) { bundle.Coverage.Floor = "" },
		"policy":   func(bundle *Bundle) { bundle.SelectionPolicy = "unknown" },
		"source":   func(bundle *Bundle) { bundle.Sources[0].SHA256 = "short" },
		"source hex": func(bundle *Bundle) {
			bundle.Sources[0].SHA256 = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
		},
		"definition": func(bundle *Bundle) { bundle.Resources[0].Definition = "missing" },
		"no definitions": func(bundle *Bundle) {
			bundle.Definitions = nil
		},
		"no resources": func(bundle *Bundle) {
			bundle.Resources = nil
		},
		"scope": func(bundle *Bundle) { bundle.Resources[0].Scope = "invalid" },
		"duplicate": func(bundle *Bundle) {
			bundle.Resources = append(bundle.Resources, bundle.Resources[0])
		},
		"order": func(bundle *Bundle) {
			bundle.Resources = append([]Resource{{APIVersion: "v1", Kind: "Pod"}}, bundle.Resources...)
		},
	}

	require.NoError(t, func() error { bundle := valid(); return bundle.Validate() }())
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			bundle := valid()
			mutate(&bundle)
			require.Error(t, bundle.Validate())
		})
	}
}
