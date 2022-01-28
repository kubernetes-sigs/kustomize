// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestExecute_Result(t *testing.T) {
	p := framework.SimpleProcessor{Filter: kio.FilterFunc(func(in []*yaml.RNode) ([]*yaml.RNode, error) {
		return in, framework.Results{
			{
				Message:  "bad value for replicas",
				Severity: framework.Error,
				ResourceRef: &yaml.ResourceIdentifier{
					TypeMeta: yaml.TypeMeta{APIVersion: "v1", Kind: "Deployment"},
					NameMeta: yaml.NameMeta{Name: "tester", Namespace: "default"},
				},
				Field: &framework.Field{
					Path:          ".spec.Replicas",
					CurrentValue:  "0",
					ProposedValue: "3",
				},
				File: &framework.File{
					Path:  "/path/to/deployment.yaml",
					Index: 0,
				},
			},
			{
				Message:  "some error",
				Severity: framework.Error,
				Tags:     map[string]string{"foo": "bar"},
			},
		}
	})}
	out := new(bytes.Buffer)
	source := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
kind: ResourceList
apiVersion: config.kubernetes.io/v1
items:
- kind: Deployment
  apiVersion: v1
  metadata:
    name: tester
    namespace: default
  spec:
    replicas: 0
`), Writer: out}
	err := framework.Execute(p, source)
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
  apiVersion: v1
  metadata:
    name: tester
    namespace: default
  spec:
    replicas: 0
results:
- message: bad value for replicas
  severity: error
  resourceRef:
    apiVersion: v1
    kind: Deployment
    name: tester
    namespace: default
  field:
    path: .spec.Replicas
    currentValue: "0"
    proposedValue: "3"
  file:
    path: /path/to/deployment.yaml
- message: some error
  severity: error
  tags:
    foo: bar`, strings.TrimSpace(out.String()))
}

const intputResourceList = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx-deployment #foo-system
  spec:
    replicas: 3
    selector:
      matchLabels:
        app: nginx
    template:
      metadata:
        labels:
          app: nginx
      spec:
        containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
          - containerPort: 80
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm
  data:
    foo: bar
`

func TestRoundtripResourceList(t *testing.T) {
	rl, err := framework.ParseResourceList([]byte(intputResourceList))
	assert.NoError(t, err)
	yml, err := rl.ToYAML()
	assert.NoError(t, err)
	assert.Equal(t, intputResourceList, yml)
}

const resourceListToSort = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: nginx-configmap
  data:
    foo: bar
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx-deployment
  spec:
    replicas: 3
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm1
  data:
    foo: bar
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm2
    namespace: staging
  data:
    foo: bar
`

const sortedResourceList = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx-deployment
  spec:
    replicas: 3
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm1
  data:
    foo: bar
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: nginx-configmap
  data:
    foo: bar
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm2
    namespace: staging
  data:
    foo: bar
`

func TestSortItems(t *testing.T) {
	rl, err := framework.ParseResourceList([]byte(resourceListToSort))
	assert.NoError(t, err)
	rl.SortItems()
	yml, err := rl.ToYAML()
	assert.NoError(t, err)
	assert.Equal(t, sortedResourceList, yml)
}
