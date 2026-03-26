// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIsLocalConfig_DefaultExcludesLocal(t *testing.T) {
	local, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-config
  annotations:
    config.kubernetes.io/local-config: "true"
`)
	assert.NoError(t, err)

	nonLocal, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: regular
`)
	assert.NoError(t, err)

	filter := IsLocalConfig{}
	result, err := filter.Filter([]*yaml.RNode{local, nonLocal})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	meta, err := result[0].GetMeta()
	assert.NoError(t, err)
	assert.Equal(t, "regular", meta.Name)
}

func TestIsLocalConfig_IncludeLocal(t *testing.T) {
	local, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-config
  annotations:
    config.kubernetes.io/local-config: "true"
`)
	assert.NoError(t, err)

	nonLocal, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: regular
`)
	assert.NoError(t, err)

	filter := IsLocalConfig{IncludeLocalConfig: true}
	result, err := filter.Filter([]*yaml.RNode{local, nonLocal})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestIsLocalConfig_ExcludeNonLocal(t *testing.T) {
	local, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-config
  annotations:
    config.kubernetes.io/local-config: "true"
`)
	assert.NoError(t, err)

	nonLocal, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: regular
`)
	assert.NoError(t, err)

	filter := IsLocalConfig{ExcludeNonLocalConfig: true}
	result, err := filter.Filter([]*yaml.RNode{local, nonLocal})
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestIsLocalConfig_IncludeLocalExcludeNonLocal(t *testing.T) {
	local, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-config
  annotations:
    config.kubernetes.io/local-config: "true"
`)
	assert.NoError(t, err)

	nonLocal, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: regular
`)
	assert.NoError(t, err)

	filter := IsLocalConfig{
		IncludeLocalConfig:    true,
		ExcludeNonLocalConfig: true,
	}
	result, err := filter.Filter([]*yaml.RNode{local, nonLocal})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	meta, err := result[0].GetMeta()
	assert.NoError(t, err)
	assert.Equal(t, "local-config", meta.Name)
}

func TestIsLocalConfig_Empty(t *testing.T) {
	filter := IsLocalConfig{}
	result, err := filter.Filter(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
