package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIsValidKubernetesResource(t *testing.T) {

	testCases := map[string]struct {
		data     yaml.ResourceIdentifier
		expected bool
	}{
		"invalid resource": {
			data: yaml.ResourceIdentifier{
				Name:       "",
				APIVersion: "",
				Kind:       "",
				Namespace:  "",
			},
			expected: false,
		},
		"valid resource": {
			data: yaml.ResourceIdentifier{
				Name:       "SomeName",
				APIVersion: "SomeVersion",
				Kind:       "SomeKind",
				Namespace:  "",
			},
			expected: true,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, IsValidKubernetesResource(tc.data), tc.expected)
		})
	}
}
