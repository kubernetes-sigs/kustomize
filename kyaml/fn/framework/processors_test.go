// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/markbates/pkger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestTemplateProcessor_ResourceTemplates(t *testing.T) {
	type API struct {
		Image string `json:"image" yaml:"image"`
	}

	p := framework.TemplateProcessor{
		TemplateData: &API{},
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.TemplatesFromDir(pkger.Dir(
				"/fn/framework/testdata/template-processor/templates/basic")),
		}},
	}

	out := new(bytes.Buffer)
	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  kind: Service
functionConfig:
  image: baz
`),
		Writer: out}

	require.NoError(t, framework.Execute(p, rw))
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

func TestTemplateProcessor_PatchTemplates(t *testing.T) {
	type API struct {
		Spec struct {
			Replicas int    `json:"replicas" yaml:"replicas"`
			A        string `json:"a" yaml:"a"`
		} `json:"spec" yaml:"spec"`
	}

	config := &API{}
	p := framework.TemplateProcessor{
		TemplateData: config,
		PatchTemplates: []framework.PatchTemplate{
			// Patch from dir with no selector templating
			&framework.ResourcePatchTemplate{
				Templates: framework.TemplatesFromDir(pkger.Dir(
					"/fn/framework/testdata/template-processor/patches/basic")),
				Selector: &framework.Selector{Names: []string{"foo"}},
			},
			// Patch from string with selector templating
			&framework.ResourcePatchTemplate{
				Selector: &framework.Selector{Names: []string{"{{.Spec.A}}"}, TemplateData: &config},
				Templates: framework.StringTemplates(`
metadata:
  annotations:
    baz: buz
`)},
		},
	}
	out := new(bytes.Buffer)

	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
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
  spec:
    replicas: 5
    a: bar
`),
		Writer: out}

	require.NoError(t, framework.Execute(p, rw))
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
        - name: foo
          image: baz
    replicas: 5
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
    annotations:
      baz: buz
  spec:
    template:
      spec:
        containers:
        - name: foo
          image: baz
functionConfig:
  spec:
    replicas: 5
    a: bar
`), strings.TrimSpace(out.String()))
}

func TestTemplateProcessor_ContainerPatchTemplates(t *testing.T) {
	type API struct {
		Spec struct {
			Key   string `json:"key" yaml:"key"`
			Value string `json:"value" yaml:"value"`
			A     string `json:"a" yaml:"a"`
		}
	}

	config := &API{}
	p := framework.TemplateProcessor{
		TemplateData: config,
		PatchTemplates: []framework.PatchTemplate{
			// patch from dir with no selector templating
			&framework.ContainerPatchTemplate{
				Templates: framework.TemplatesFromDir(pkger.Dir(
					"/fn/framework/testdata/template-processor/container-patches")),
				Selector: &framework.Selector{Names: []string{"foo"}},
			},
			// patch from string with selector templating
			&framework.ContainerPatchTemplate{
				Selector: &framework.Selector{Names: []string{"{{.Spec.A}}"}, TemplateData: &config},
				Templates: framework.StringTemplates(`
env:
- name: Foo
  value: Bar
`)},
		},
	}

	out := new(bytes.Buffer)
	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
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
          - name: EXISTING
            value: variable
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
  spec:
    key: Hello
    value: World
    a: bar
`),
		Writer: out}

	require.NoError(t, framework.Execute(p, rw))
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
          - name: EXISTING
            value: variable
          - name: Hello
            value: World
        - name: b
          env:
          - name: Hello
            value: World
        - name: c
          env:
          - name: Hello
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
          env:
          - name: Foo
            value: Bar
functionConfig:
  spec:
    key: Hello
    value: World
    a: bar
`), strings.TrimSpace(out.String()))
}

func TestSimpleProcessor_Process_loads_config(t *testing.T) {
	cfg := new(struct {
		Value string `yaml:"value"`
	})
	p := framework.SimpleProcessor{
		Filter: kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
			if cfg.Value != "dataFromResourceList" {
				return nil, errors.Errorf("got incorrect config value %q", cfg.Value)
			}
			return items, nil
		}),
		Config: &cfg,
	}
	rl := framework.ResourceList{
		FunctionConfig: yaml.NewMapRNode(&map[string]string{
			"value": "dataFromResourceList",
		}),
	}
	require.NoError(t, p.Process(&rl))
}

func TestSimpleProcessor_Process_Error(t *testing.T) {
	tests := []struct {
		name    string
		filter  kio.Filter
		config  interface{}
		wantErr string
	}{
		{
			name:    "error when given func as Config",
			config:  func() {},
			wantErr: "cannot unmarshal !!map into func()",
		},
		{
			name:    "error in filter",
			wantErr: "err from filter",
			filter: kio.FilterFunc(func(_ []*yaml.RNode) ([]*yaml.RNode, error) {
				return nil, errors.Errorf("err from filter")
			}),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := framework.SimpleProcessor{
				Filter: tt.filter,
				Config: tt.config,
			}
			rl := framework.ResourceList{
				FunctionConfig: yaml.NewMapRNode(&map[string]string{
					"value": "dataFromResourceList",
				}),
			}
			err := p.Process(&rl)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestVersionedAPIProcessor_Process_Error(t *testing.T) {
	tests := []struct {
		name           string
		filterProvider framework.FilterProvider
		apiVersion     string
		kind           string
		wantErr        string
	}{
		{
			name: "error when given FilterFunc as Filter",
			filterProvider: framework.FilterProviderFunc(func(_, _ string) (kio.Filter, error) {
				return kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
					return items, nil
				}), nil
			}),
			wantErr: "cannot unmarshal !!map into kio.FilterFunc",
		},
		{
			name: "error in filter",
			filterProvider: framework.FilterProviderFunc(func(_, _ string) (kio.Filter, error) {
				return &framework.AndSelector{FailOnEmptyMatch: true}, nil
			}),
			wantErr: "selector did not select any items",
		},
		{
			name: "error GVKFilterMap no filter for kind",
			filterProvider: framework.GVKFilterMap{
				"puppy": {
					"pets.example.com/v1beta1": &framework.Selector{},
				},
			},
			kind:       "kitten",
			apiVersion: "pets.example.com/v1beta1",
			wantErr:    "kind \"kitten\" is not supported",
		},
		{
			name: "error GVKFilterMap no filter for version",
			filterProvider: framework.GVKFilterMap{
				"kitten": {
					"pets.example.com/v1alpha1": &framework.Selector{},
				},
			},
			kind:       "kitten",
			apiVersion: "pets.example.com/v1beta1",
			wantErr:    "apiVersion \"pets.example.com/v1beta1\" is not supported for kind \"kitten\"",
		},
		{
			name:           "error GVKFilterMap blank kind",
			filterProvider: framework.GVKFilterMap{},
			kind:           "",
			apiVersion:     "pets.example.com/v1beta1",
			wantErr:        "unable to identify provider for resource: kind is required",
		},
		{
			name:           "error GVKFilterMap blank version",
			filterProvider: framework.GVKFilterMap{},
			kind:           "kitten",
			apiVersion:     "",
			wantErr:        "unable to identify provider for resource: apiVersion is required",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := framework.VersionedAPIProcessor{
				FilterProvider: tt.filterProvider,
			}
			rl := framework.ResourceList{
				FunctionConfig: yaml.NewMapRNode(&map[string]string{
					"apiVersion": tt.apiVersion,
					"kind":       tt.kind,
				}),
			}
			err := p.Process(&rl)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestTemplateProcessor_Process_Error(t *testing.T) {
	tests := []struct {
		name      string
		processor framework.TemplateProcessor
		wantErr   string
	}{
		{
			name: "ResourcePatchTemplate is not a resource",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ResourcePatchTemplate{
						Templates: framework.StringTemplates(`aString
another`),
					}},
			},
			wantErr: `failed to parse rendered patch template into a resource:
001 aString
002 another
: wrong Node Kind for  expected: MappingNode was ScalarNode: value: {aString another}`,
		},
		{
			name: "ResourcePatchTemplate is invalid template",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ResourcePatchTemplate{
						Templates: framework.StringTemplates("foo: {{ .OOPS }}"),
					}},
			},
			wantErr: "can't evaluate field OOPS",
		},
		{
			name: "ContainerPatchTemplate is not a resource",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ContainerPatchTemplate{
						Templates: framework.StringTemplates(`aString
another`),
					}},
			},
			wantErr: `failed to parse rendered patch template into a resource:
001 aString
002 another
: wrong Node Kind for  expected: MappingNode was ScalarNode: value: {aString another}`,
		},
		{
			name: "ContainerPatchTemplate is invalid template",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ContainerPatchTemplate{
						Templates: framework.StringTemplates("foo: {{ .OOPS }}"),
					}},
			},
			wantErr: "can't evaluate field OOPS",
		},
		{
			name: "ResourceTemplate is not a resource",
			processor: framework.TemplateProcessor{
				ResourceTemplates: []framework.ResourceTemplate{{
					Templates: framework.StringTemplates(`aString
another`),
				}},
			},
			wantErr: `failed to parse rendered template into a resource:
001 aString
002 another
: wrong Node Kind for  expected: MappingNode was ScalarNode: value: {aString another}`,
		},
		{
			name: "ResourceTemplate is invalid template",
			processor: framework.TemplateProcessor{
				ResourceTemplates: []framework.ResourceTemplate{{
					Templates: framework.StringTemplates("foo: {{ .OOPS }}"),
				}},
			},
			wantErr: "can't evaluate field OOPS",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rl := framework.ResourceList{
				Items: []*yaml.RNode{
					yaml.MustParse(`
kind: Deployment
apiVersion: apps/v1
metadata:
  name: foo
spec:
  replicas: 5
  template:
    spec:
      containers:
      - name: foo
`),
				},
				FunctionConfig: yaml.NewMapRNode(&map[string]string{
					"value": "dataFromResourceList",
				}),
			}
			tt.processor.TemplateData = new(struct {
				Value string `yaml:"value"`
			})
			err := tt.processor.Process(&rl)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestTemplateProcessor_AdditionalSchemas(t *testing.T) {
	p := framework.TemplateProcessor{
		AdditionalSchemas: func() ([]*spec.Definitions, error) {
			// This adds the same thing twice, just to exercise both the ...FromDir and the ...FromFile helpers
			c1, err := framework.SchemaDefinitionsFromDir("/fn/framework/testdata/template-processor/schemas")()
			if err != nil {
				return nil, errors.WrapPrefixf(err, "schema from dir")
			}
			c2, err := framework.SchemaDefinitionsFromFile("/fn/framework/testdata/template-processor/schemas/foo.json")()
			if err != nil {
				return nil, errors.WrapPrefixf(err, "schema from file")
			}
			return append(c1, c2...), nil
		},
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.TemplatesFromFile("/fn/framework/testdata/template-processor/templates/custom-resource/foo.yaml"),
		}},
		PatchTemplates: []framework.PatchTemplate{
			&framework.ResourcePatchTemplate{
				Templates: framework.TemplatesFromFile("/fn/framework/testdata/template-processor/patches/custom-resource/patch.template.yaml")},
		},
	}
	out := new(bytes.Buffer)

	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: example.com/v1
  kind: Foo
  metadata:
    name: source
  spec:
    targets:
    - app: C
      size: medium
`),
		Writer: out}
	defer openapi.ResetOpenAPI()
	require.NoError(t, framework.Execute(p, rw))
	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: example.com/v1
  kind: Foo
  metadata:
    name: source
  spec:
    targets:
    - app: C
      size: large
      type: Ruby
    - app: B
      size: small
- apiVersion: example.com/v1
  kind: Foo
  metadata:
    name: example
  spec:
    targets:
    - app: A
      type: Go
      size: small
    - app: B
      type: Go
      size: small
    - app: C
      type: Ruby
      size: large
`), strings.TrimSpace(out.String()))
}
