// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func mustParseNode(t *testing.T, value string) *yaml.RNode {
	t.Helper()
	node, err := yaml.Parse(value)
	require.NoError(t, err)
	return node
}

func TestContainerNameMatcher(t *testing.T) {
	container := mustParseNode(t, `
name: app
image: example.io/app:v1
`)

	require.True(t, framework.ContainerNameMatcher()(container))
	require.True(t, framework.ContainerNameMatcher("sidecar", "app")(container))
	require.False(t, framework.ContainerNameMatcher("sidecar")(container))

	noName := mustParseNode(t, `image: example.io/app:v1`)
	require.False(t, framework.ContainerNameMatcher("app")(noName))
	require.True(t, framework.ContainerNameMatcher()(noName))
}

func TestLabelMatcher(t *testing.T) {
	node := mustParseNode(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
  labels:
    app: foo
    env: prod
`)

	m := framework.LabelMatcher(map[string]string{"app": "foo"})
	require.NoError(t, m.InitTemplates())
	require.True(t, m.Match(node))

	m = framework.LabelMatcher(map[string]string{"app": "foo", "env": "staging"})
	require.NoError(t, m.InitTemplates())
	require.False(t, m.Match(node))

	m = framework.LabelMatcher(map[string]string{"missing": "key"})
	require.NoError(t, m.InitTemplates())
	require.False(t, m.Match(node))

	m = framework.LabelMatcher(map[string]string{})
	require.NoError(t, m.InitTemplates())
	require.True(t, m.Match(node))
}

func TestAnnotationMatcher(t *testing.T) {
	node := mustParseNode(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
  annotations:
    example.io/select: "yes"
`)

	m := framework.AnnotationMatcher(map[string]string{"example.io/select": "yes"})
	require.NoError(t, m.InitTemplates())
	require.True(t, m.Match(node))

	m = framework.AnnotationMatcher(map[string]string{"example.io/select": "no"})
	require.NoError(t, m.InitTemplates())
	require.False(t, m.Match(node))
}

func TestAPIVersionMatcher(t *testing.T) {
	node := mustParseNode(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
`)

	m := framework.APIVersionMatcher("apps/v1")
	require.NoError(t, m.InitTemplates())
	require.True(t, m.Match(node))

	m = framework.APIVersionMatcher("apps/v2")
	require.NoError(t, m.InitTemplates())
	require.False(t, m.Match(node))
}

func TestGVKMatcher(t *testing.T) {
	node := mustParseNode(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
`)

	m := framework.GVKMatcher("apps/v1/Deployment")
	require.NoError(t, m.InitTemplates())
	require.True(t, m.Match(node))

	m = framework.GVKMatcher("apps/v1/StatefulSet")
	require.NoError(t, m.InitTemplates())
	require.False(t, m.Match(node))
}
