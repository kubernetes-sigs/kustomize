// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/testutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCommand_dockerfile(t *testing.T) {
	d, err := ioutil.TempDir("", "kustomize")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(d)

	// create a function

	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error { return nil })

	// generate the Dockerfile
	cmd.SetArgs([]string{"gen", d})
	if !assert.NoError(t, cmd.Execute()) {
		t.FailNow()
	}

	b, err := ioutil.ReadFile(filepath.Join(d, "Dockerfile"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := `FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
`
	if !assert.Equal(t, expected, string(b)) {
		t.FailNow()
	}
}

// TestCommand_standalone tests the framework works in standalone mode
func TestCommand_standalone(t *testing.T) {
	// TODO: make this test pass on windows -- currently failure seems spurious
	testutil.SkipWindows(t)

	type api = struct {
		A string `json:"a" yaml:"a"`
	}
	var config api

	resourceList := &framework.ResourceList{FunctionConfig: &config}
	cmdFn := func() *cobra.Command {
		return framework.Command(resourceList, func() error {
			resourceList.Items = append(resourceList.Items, yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
`))
			for i := range resourceList.Items {
				err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a", config.A))
				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	frameworktestutil.ResultsChecker{Command: cmdFn}.Assert(t)
}

func TestCommand_standalonestdin(t *testing.T) {
	// TODO: make this test pass on windows -- currently failure seems spurious
	testutil.SkipWindows(t)

	type api = struct {
		A string `json:"a" yaml:"a"`
	}
	var config api

	resourceList := &framework.ResourceList{FunctionConfig: &config}
	cmd := framework.Command(resourceList, func() error {
		resourceList.Items = append(resourceList.Items, yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar2
  namespace: default
  annotations:
    foo: bar2
`))
		for i := range resourceList.Items {
			err := resourceList.Items[i].PipeE(yaml.SetAnnotation("a", config.A))
			if err != nil {
				return err
			}
		}

		return nil
	})
	cmd.SetIn(bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
spec:
  replicas: 1
`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{filepath.Join("testdata", "command", "config.yaml"), "-"})

	require.NoError(t, cmd.Execute())

	require.Equal(t, strings.TrimSpace(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar1
  namespace: default
  annotations:
    foo: bar1
    a: 'b'
spec:
  replicas: 1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar2
  namespace: default
  annotations:
    foo: bar2
    a: 'b'
`), strings.TrimSpace(out.String()))
}

func TestCommand_PatchTemplateFn(t *testing.T) {
	// TODO: make this test pass on windows -- currently failure seems spurious
	testutil.SkipWindows(t)

	type api = struct {
		Spec struct {
			A string `json:"a" yaml:"a"`
		} `json:"spec" yaml:"spec"`
	}
	var config api

	cmd := framework.TemplateCommand{
		API: &config,
		PatchTemplatesFn: func(_ *framework.ResourceList) ([]framework.PatchTemplate, error) {
			return []framework.PatchTemplate{{
				Selector: &framework.Selector{Names: []string{config.Spec.A}},
				Template: template.Must(template.New("test").Parse(`
metadata:
  annotations:
    baz: buz
`)),
			}}, nil
		},
	}.GetCommand()

	cmd.SetIn(bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar1
    namespace: default
    annotations:
      foo: bar1
  spec:
    replicas: 1
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar2
    namespace: default
    annotations:
      foo: bar2
  spec:
    replicas: 1
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: Example
  spec:
    a: "bar1"
`))
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, cmd.Execute())

	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar1
    namespace: default
    annotations:
      foo: bar1
      baz: buz
      config.kubernetes.io/index: '0'
  spec:
    replicas: 1
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar2
    namespace: default
    annotations:
      foo: bar2
  spec:
    replicas: 1
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: Example
  spec:
    a: "bar1"
`), strings.TrimSpace(out.String()))
}

func TestCommand_PatchContainerTemplatesFn(t *testing.T) {
	// TODO: make this test pass on windows -- currently failure seems spurious
	testutil.SkipWindows(t)

	type api = struct {
		Spec struct {
			A string `json:"a" yaml:"a"`
		} `json:"spec" yaml:"spec"`
	}
	var config api

	cmd := framework.TemplateCommand{
		API: &config,
		PatchContainerTemplatesFn: func(_ *framework.ResourceList) ([]framework.ContainerPatchTemplate, error) {
			return []framework.ContainerPatchTemplate{{
				PatchTemplate: framework.PatchTemplate{
					Selector: &framework.Selector{Names: []string{config.Spec.A}},
					Template: template.Must(template.New("test").Parse(`
env:
  key: Foo
  value: Bar
`))},
			}}, nil
		},
	}.GetCommand()

	cmd.SetIn(bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar1
    namespace: default
    annotations:
      foo: bar1
  spec:
    template:
      spec:
        containers:
        - name: foo
        - name: bar
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar2
    namespace: default
    annotations:
      foo: bar2
  spec:
    template:
      spec:
        containers:
        - name: foo
        - name: bar
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: Example
  spec:
    a: "bar1"
`))
	var out bytes.Buffer
	cmd.SetOut(&out)

	require.NoError(t, cmd.Execute())

	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar1
    namespace: default
    annotations:
      foo: bar1
  spec:
    template:
      spec:
        containers:
        - name: foo
          env:
            key: Foo
            value: Bar
        - name: bar
          env:
            key: Foo
            value: Bar
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar2
    namespace: default
    annotations:
      foo: bar2
  spec:
    template:
      spec:
        containers:
        - name: foo
        - name: bar
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: Example
  spec:
    a: "bar1"
`), strings.TrimSpace(out.String()))
}
