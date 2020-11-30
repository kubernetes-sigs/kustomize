// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run github.com/markbates/pkger/cmd/pkger -o fn/framework/example2
package example2

import (
	"bytes"
	"strings"
	"testing"

	"github.com/markbates/pkger"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func TestTemplate(t *testing.T) {
	type API struct {
		Image string `json:"image" yaml:"image"`
	}

	tpl, err := framework.TemplatesFromDir(pkger.Dir("/fn/framework/example2/data/templates"))(nil)
	require.NoError(t, err)
	cmd := framework.TemplateCommand{
		API:       &API{},
		Templates: tpl,
	}.GetCommand()

	var in, out bytes.Buffer
	cmd.SetIn(&in)
	cmd.SetOut(&out)

	in.WriteString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
functionConfig:
  image: baz
`)

	require.NoError(t, cmd.Execute())
	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
    namespace: bar
    annotations:
      config.kubernetes.io/index: '0'
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  image: baz
`), strings.TrimSpace(out.String()))
}

func TestPatchTemplate(t *testing.T) {
	type API struct {
		Replicas int `json:"replicas" yaml:"replicas"`
	}

	cmd := framework.TemplateCommand{
		API: &API{},
		PatchTemplatesFn: framework.PatchTemplatesFromDir(
			framework.PT{
				Dir: pkger.Dir("/fn/framework/example2/data/patches"),
				Selector: func() *framework.Selector {
					return &framework.Selector{Names: []string{"foo"}}
				},
			},
		),
	}.GetCommand()

	var in, out bytes.Buffer
	cmd.SetIn(&in)
	cmd.SetOut(&out)

	in.WriteString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  replicas: 5
`)

	require.NoError(t, cmd.Execute())
	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
    annotations:
      config.kubernetes.io/index: '0'
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
    replicas: 5
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  replicas: 5
`), strings.TrimSpace(out.String()))
}

func TestContainerPatchTemplate(t *testing.T) {
	type API struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	}

	cmd := framework.TemplateCommand{
		API: &API{},
		PatchContainerTemplatesFn: framework.ContainerPatchTemplatesFromDir(
			framework.CPT{
				Dir: pkger.Dir("/fn/framework/example2/data/container-patches"),
				Selector: func() *framework.Selector {
					return &framework.Selector{Names: []string{"foo"}}
				},
			},
		),
	}.GetCommand()

	var in, out bytes.Buffer
	cmd.SetIn(&in)
	cmd.SetOut(&out)

	in.WriteString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
  spec:
    template:
      spec:
        containers:
        - name: a
        - name: b
        - name: c
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  key: Hello
  value: World
`)

	require.NoError(t, cmd.Execute())
	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
  spec:
    template:
      spec:
        containers:
        - name: a
          env:
            key: Hello
            value: World
        - name: b
          env:
            key: Hello
            value: World
        - name: c
          env:
            key: Hello
            value: World
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  key: Hello
  value: World
`), strings.TrimSpace(out.String()))
}
