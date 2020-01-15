package cmd

import (
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"testing"
)

func TestIsValidKubernetesResource(t *testing.T) {
	assert.False(t, IsValidKubernetesResource(nil))

	id := yaml.ResourceIdentifier{
		Name : "",
		APIVersion : "",
		Kind : "",
		Namespace : "",
	}
	assert.False(t, IsValidKubernetesResource(&id))

	id.Name = "SomeName"
	id.APIVersion = "SomeVersion"
	id.Kind = "SomeKind"
	assert.True(t, IsValidKubernetesResource(&id))

	id.APIVersion = ""
	assert.False(t, IsValidKubernetesResource(&id))
}
