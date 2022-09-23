// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	validationErrors "k8s.io/kube-openapi/pkg/validation/errors"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/resid"
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
			Templates: parser.TemplateFiles("testdata/template-processor/templates/basic"),
		}},
	}

	out := new(bytes.Buffer)
	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
				Templates: parser.TemplateFiles("testdata/template-processor/patches/basic"),
				Selector:  &framework.Selector{Names: []string{"foo"}},
			},
			// Patch from string with selector templating
			&framework.ResourcePatchTemplate{
				Selector: &framework.Selector{Names: []string{"{{.Spec.A}}"}, TemplateData: &config},
				Templates: parser.TemplateStrings(`
metadata:
  annotations:
    baz: buz
`)},
		},
	}
	out := new(bytes.Buffer)

	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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
				Templates: parser.TemplateFiles("testdata/template-processor/container-patches"),
				Selector:  &framework.Selector{Names: []string{"foo"}},
			},
			// patch from string with selector templating
			&framework.ContainerPatchTemplate{
				Selector: &framework.Selector{Names: []string{"{{.Spec.A}}"}, TemplateData: &config},
				Templates: parser.TemplateStrings(`
env:
- name: Foo
  value: Bar
`)},
		},
	}

	out := new(bytes.Buffer)
	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1
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
apiVersion: config.kubernetes.io/v1
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

func TestTemplateProcessor_ContainerPatchTemplates_MultipleWorkloadKinds(t *testing.T) {
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
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: parser.TemplateFiles("testdata/template-processor/templates/container-sources"),
		}},
		PatchTemplates: []framework.PatchTemplate{
			&framework.ContainerPatchTemplate{
				Templates: parser.TemplateFiles("testdata/template-processor/container-patches"),
			},
		},
	}

	out := new(bytes.Buffer)
	rw := &kio.ByteReadWriter{Writer: out, Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items: []
functionConfig:
  spec:
    key: Hello
    value: World
    a: bar
`)}

	require.NoError(t, framework.Execute(p, rw))
	resources, err := (&kio.ByteReader{Reader: out}).Read()
	require.NoError(t, err)
	envRegex := regexp.MustCompile(strings.TrimSpace(`
\s+ env:
\s+ - name: EXISTING
\s+   value: variable
\s+ - name: Hello
\s+   value: World
`))
	require.Equal(t, 9, len(resources))
	for i, r := range resources {
		t.Run(r.GetKind(), func(t *testing.T) {
			assert.Regexp(t, envRegex, resources[i].MustString())
		})
	}
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
			name:    "error when filter is nil",
			config:  map[string]string{},
			filter:  nil,
			wantErr: "processing filter: ResourceList cannot run apply nil filter",
		}, {
			name:   "no error when config is nil",
			config: nil,
			filter: kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
				return items, nil
			}),
			wantErr: "",
		},
		{
			name:    "error in filter",
			wantErr: "processing filter: err from filter",
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
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
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
						Templates: parser.TemplateStrings(`aString
another`),
					}},
			},
			wantErr: `failed to parse rendered patch template into a resource:
001 aString
002 another
: wrong node kind: expected MappingNode but got ScalarNode: node contents:
aString another
`,
		},
		{
			name: "ResourcePatchTemplate is invalid template",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ResourcePatchTemplate{
						Templates: parser.TemplateStrings("foo: {{ .OOPS }}"),
					}},
			},
			wantErr: "can't evaluate field OOPS",
		},
		{
			name: "ContainerPatchTemplate is not a resource",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ContainerPatchTemplate{
						Templates: parser.TemplateStrings(`aString
another`),
					}},
			},
			wantErr: `failed to parse rendered patch template into a resource:
001 aString
002 another
: wrong node kind: expected MappingNode but got ScalarNode: node contents:
aString another
`,
		},
		{
			name: "ContainerPatchTemplate is invalid template",
			processor: framework.TemplateProcessor{
				PatchTemplates: []framework.PatchTemplate{
					&framework.ContainerPatchTemplate{
						Templates: parser.TemplateStrings("foo: {{ .OOPS }}"),
					}},
			},
			wantErr: "can't evaluate field OOPS",
		},
		{
			name: "ResourceTemplate is not a resource",
			processor: framework.TemplateProcessor{
				ResourceTemplates: []framework.ResourceTemplate{{
					Templates: parser.TemplateStrings(`aString
another`),
				}},
			},
			wantErr: `failed to parse rendered template into a resource:
001 aString
002 another
: wrong node kind: expected MappingNode but got ScalarNode: node contents:
aString another
`,
		},
		{
			name: "ResourceTemplate is invalid template",
			processor: framework.TemplateProcessor{
				ResourceTemplates: []framework.ResourceTemplate{{
					Templates: parser.TemplateStrings("foo: {{ .OOPS }}"),
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
		AdditionalSchemas: parser.SchemaFiles("testdata/template-processor/schemas"),
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: parser.TemplateFiles("testdata/template-processor/templates/custom-resource/foo.template.yaml"),
		}},
		PatchTemplates: []framework.PatchTemplate{
			&framework.ResourcePatchTemplate{
				Templates: parser.TemplateFiles("testdata/template-processor/patches/custom-resource/patch.template.yaml")},
		},
	}
	out := new(bytes.Buffer)

	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1
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
	require.NoError(t, framework.Execute(p, rw))
	require.Equal(t, strings.TrimSpace(`
apiVersion: config.kubernetes.io/v1
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
	found := openapi.SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "example.com/v1",
		Kind:       "Foo",
	})
	require.Nil(t, found, "openAPI schema was not reset")
}

func TestTemplateProcessor_Validator(t *testing.T) {
	// This test proves the Validate method is called when implemented
	// and demonstrates the use of ProcessorResultsChecker's error matching
	p := func() framework.ResourceListProcessor {
		return &framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
			"JavaSpringBoot": {
				"example.com/v1alpha1": &v1alpha1JavaSpringBoot{},
			}}}
	}
	c := frameworktestutil.ProcessorResultsChecker{
		TestDataDirectory: "testdata/validation",
		Processor:         p,
	}
	c.Assert(t)
}

type jsonTagTest struct {
	Name string `json:"name"`
	Test bool   `json:"test"`
}

type yamlTagTest struct {
	Name string `yaml:"name"`
	Test bool   `yaml:"test"`
}

type customErrorTest struct {
	v1alpha1JavaSpringBoot
}

func (e customErrorTest) Schema() (*spec.Schema, error) {
	return e.v1alpha1JavaSpringBoot.Schema()
}

func (e customErrorTest) Validate() error {
	return errors.Errorf("Custom errors:\n- first error\n- second error")
}

type errorMergeTest struct {
	v1alpha1JavaSpringBoot
}

func (e errorMergeTest) Schema() (*spec.Schema, error) {
	return e.v1alpha1JavaSpringBoot.Schema()
}

func (e errorMergeTest) Validate() error {
	if strings.HasSuffix(e.Spec.Image, "latest") {
		return validationErrors.CompositeValidationError(errors.Errorf("spec.image cannot be tagged :latest"))
	}
	return nil
}

func (a schemaProviderOnlyTest) Schema() (*spec.Schema, error) {
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("example.com", "v1alpha1", "JavaSpringBoot"), javaSpringBootDefinition)
	return schema, errors.WrapPrefixf(err, "parsing JavaSpringBoot schema")
}

type schemaProviderOnlyTest struct {
	Metadata Metadata                   `yaml:"metadata" json:"metadata"`
	Spec     v1alpha1JavaSpringBootSpec `yaml:"spec" json:"spec"`
}

func TestLoadFunctionConfig(t *testing.T) {
	tests := []struct {
		name        string
		src         *yaml.RNode
		api         interface{}
		want        interface{}
		wantErrMsgs []string
	}{
		{
			name: "combines schema-based and non-composite custom errors",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1 
kind: JavaSpringBoot
spec:
  replicas: 11
  domain: foo.myco.io
  image: nginx:latest
`),
			api: &customErrorTest{},
			wantErrMsgs: []string{
				"validation failure list:",
				"spec.replicas in body should be less than or equal to 9",
				"spec.domain in body should match 'example\\.com$'",
				`Custom errors:
- first error
- second error`,
			},
		},
		{
			name: "merges schema-based errors with custom composite errors",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1 
kind: JavaSpringBoot
spec:
  replicas: 11
  domain: foo.myco.io
  image: nginx:latest
`),
			api: &errorMergeTest{},
			wantErrMsgs: []string{"validation failure list:",
				"spec.replicas in body should be less than or equal to 9",
				"spec.domain in body should match 'example\\.com$'",
				"spec.image cannot be tagged :latest"},
		},
		{
			name: "schema errors only",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1 
kind: JavaSpringBoot
spec:
  replicas: 11
`),
			api: &errorMergeTest{},
			wantErrMsgs: []string{
				`validation failure list:
spec.replicas in body should be less than or equal to 9`,
			},
		}, {
			name: "schema provider only",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1
kind: JavaSpringBoot
spec:
  replicas: 11
`),
			api: &schemaProviderOnlyTest{},
			wantErrMsgs: []string{
				`validation failure list:
spec.replicas in body should be less than or equal to 9`,
			},
		},
		{
			name: "custom errors only",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1 
kind: JavaSpringBoot
spec:
  image: nginx:latest
`),
			api: &errorMergeTest{},
			wantErrMsgs: []string{
				`validation failure list:
spec.image cannot be tagged :latest`},
		},
		{
			name: "both custom and schema error hooks defined, but no errors produced",
			src: yaml.MustParse(`
apiVersion: example.com/v1alpha1 
kind: JavaSpringBoot
spec:
  image: nginx:1.0
  replicas: 3
  domain: bar.example.com
`),
			api: &errorMergeTest{},
			want: &errorMergeTest{v1alpha1JavaSpringBoot: v1alpha1JavaSpringBoot{
				Spec: v1alpha1JavaSpringBootSpec{Replicas: 3, Domain: "bar.example.com", Image: "nginx:1.0"}},
			},
		},
		{
			name: "successfully loads types that include fields only tagged with json markers",
			src: yaml.MustParse(`
name: tester
test: true
`),
			api:  &jsonTagTest{},
			want: &jsonTagTest{Name: "tester", Test: true},
		},
		{
			name: "successfully loads types that include fields only tagged with yaml markers",
			src: yaml.MustParse(`
name: tester
test: true
`),
			api:  &yamlTagTest{},
			want: &yamlTagTest{Name: "tester", Test: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := framework.LoadFunctionConfig(tt.src, tt.api)
			if len(tt.wantErrMsgs) == 0 {
				require.NoError(t, err)
				require.Equal(t, tt.want, tt.api)
			} else {
				require.Error(t, err)
				for _, msg := range tt.wantErrMsgs {
					require.Contains(t, err.Error(), msg)
				}
			}
		})
	}
}
