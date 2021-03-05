// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const service = "Service"

// ExampleSimpleProcessor_modify implements a function that sets an annotation on each resource.
func ExampleSimpleProcessor_modify() {
	input := bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
# items are provided as nodes
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
- apiVersion: v1
  kind: Service
  metadata:
    name: foo
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  data:
    value: baz
`)
	config := new(struct {
		Data map[string]string `yaml:"data" json:"data"`
	})
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
			if err := items[i].PipeE(yaml.SetAnnotation("value", config.Data["value"])); err != nil {
				return nil, err
			}
		}
		return items, nil
	}

	err := framework.Execute(framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}, &kio.ByteReadWriter{Reader: input})
	if err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1alpha1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: foo
	//     annotations:
	//       value: 'baz'
	// - apiVersion: v1
	//   kind: Service
	//   metadata:
	//     name: foo
	//     annotations:
	//       value: 'baz'
	// functionConfig:
	//   apiVersion: v1
	//   kind: ConfigMap
	//   data:
	//     value: baz
}

// ExampleSimpleProcessor_generateReplace generates a resource from a FunctionConfig.
// If the resource already exists, it replaces the resource with a new copy.
func ExampleSimpleProcessor_generateReplace() {
	input := bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
# items are provided as nodes
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: ExampleServiceGenerator
  spec:
    name: bar
`)

	// function API definition which will be parsed from the ResourceList.FunctionConfig
	// read from stdin
	type Spec struct {
		Name string `yaml:"name,omitempty"`
	}
	type ExampleServiceGenerator struct {
		Spec Spec `yaml:"spec,omitempty"`
	}

	functionConfig := &ExampleServiceGenerator{}

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		// remove the last generated resource
		var newNodes []*yaml.RNode
		for i := range items {
			meta, err := items[i].GetMeta()
			if err != nil {
				return nil, err
			}
			// something we already generated, remove it from the list so we regenerate it
			if meta.Name == functionConfig.Spec.Name &&
				meta.Kind == service &&
				meta.APIVersion == "v1" {
				continue
			}
			newNodes = append(newNodes, items[i])
		}
		items = newNodes

		// generate the resource again
		n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
 name: %s
`, functionConfig.Spec.Name))
		if err != nil {
			return nil, err
		}
		items = append(items, n)
		return items, nil
	}

	p := framework.SimpleProcessor{Config: functionConfig, Filter: kio.FilterFunc(fn)}
	err := framework.Execute(p, &kio.ByteReadWriter{Reader: input})
	if err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1alpha1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: foo
	// - apiVersion: v1
	//   kind: Service
	//   metadata:
	//     name: bar
	// functionConfig:
	//   apiVersion: example.com/v1alpha1
	//   kind: ExampleServiceGenerator
	//   spec:
	//     name: bar
}

// ExampleTemplateProcessor provides an example for using the TemplateProcessor to add resources
// from templates defined inline
func ExampleTemplateProcessor_generate_inline() {
	api := new(struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	})
	// create the template
	fn := framework.TemplateProcessor{
		// Templates input
		TemplateData: api,
		// Templates
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: foo
 namespace: default
 annotations:
   {{ .Key }}: {{ .Value }}
`)}},
	}
	cmd := command.Build(fn, command.StandaloneEnabled, false)

	// mimic standalone mode: testdata/template/config.yaml will be parsed into `api`
	cmd.SetArgs([]string{filepath.Join("testdata", "example", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: default
	//   annotations:
	//     a: b
}

// ExampleTemplateProcessor_files provides an example for using the TemplateProcessor to add
// resources from templates defined in files.
func ExampleTemplateProcessor_generate_files() {
	api := new(struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	})
	// create the template
	templateFn := framework.TemplateProcessor{
		// Templates input
		TemplateData: api,
		// Templates
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.TemplatesFromFile(
				filepath.Join("testdata", "example", "templatefiles", "deployment.template"),
			),
		}},
	}
	cmd := command.Build(templateFn, command.StandaloneEnabled, false)
	// mimic standalone mode: testdata/template/config.yaml will be parsed into `api`
	cmd.SetArgs([]string{filepath.Join("testdata", "example", "templatefiles", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: default
	//   annotations:
	//     a: b
}

// ExampleTemplateProcessor_preprocess provides an example for using the TemplateProcessor
// with PreProcess to configure the template based on the input resources observed.
func ExampleTemplateProcessor_preprocess() {
	config := new(struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
		Short bool
	})

	// create the template
	fn := framework.TemplateProcessor{
		// Templates input
		TemplateData: config,
		PreProcessFilters: []kio.Filter{
			kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
				config.Short = len(items) < 3
				return items, nil
			}),
		},
		// Templates
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
 name: foo
 namespace: default
 annotations:
   {{ .Key }}: {{ .Value }}
{{- if .Short }}
   short: 'true'
{{- end }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
 name: bar
 namespace: default
 annotations:
   {{ .Key }}: {{ .Value }}
{{- if .Short }}
   short: 'true'
{{- end }}
`),
		}},
	}

	cmd := command.Build(fn, command.StandaloneEnabled, false)
	// mimic standalone mode: testdata/template/config.yaml will be parsed into `api`
	cmd.SetArgs([]string{filepath.Join("testdata", "example", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: default
	//   annotations:
	//     a: b
	//     short: 'true'
	// ---
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	//   namespace: default
	//   annotations:
	//     a: b
	//     short: 'true'
}

// ExampleTemplateProcessor_postprocess provides an example for using the TemplateProcessor
// with PostProcess to modify the results.
func ExampleTemplateProcessor_postprocess() {
	config := new(struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	})

	// create the template
	fn := framework.TemplateProcessor{
		// Templates input
		TemplateData: config,
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
`),
		}},
		PostProcessFilters: []kio.Filter{
			kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
				items = items[1:]
				return items, nil
			}),
		},
	}
	cmd := command.Build(fn, command.StandaloneEnabled, false)

	cmd.SetArgs([]string{filepath.Join("testdata", "example", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	//   namespace: default
	//   annotations:
	//     a: b
}

// ExampleTemplateProcessor_patch provides an example for using the TemplateProcessor to
// create a function that patches resources.
func ExampleTemplateProcessor_patch() {
	fn := framework.TemplateProcessor{
		TemplateData: new(struct {
			Key   string `json:"key" yaml:"key"`
			Value string `json:"value" yaml:"value"`
		}),
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
`),
		}},
		// PatchTemplates are applied to BOTH ResourceList input resources AND templated resources
		PatchTemplates: []framework.PatchTemplate{
			&framework.ResourcePatchTemplate{
				// patch the foo resource only
				Selector: &framework.Selector{Names: []string{"foo"}},
				Templates: framework.StringTemplates(`
metadata:
  annotations:
    patched: 'true'
`),
			}},
	}
	cmd := command.Build(fn, command.StandaloneEnabled, false)

	cmd.SetArgs([]string{filepath.Join("testdata", "example", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: default
	//   annotations:
	//     a: b
	//     patched: 'true'
	// ---
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	//   namespace: default
	//   annotations:
	//     a: b
}

// ExampleTemplateProcessor_MergeResources provides an example for using the TemplateProcessor to
// create a function that treats incoming resources as potential patches
// for the resources the function generates itself.
func ExampleTemplateProcessor_MergeResources() {
	p := framework.TemplateProcessor{
		TemplateData: new(struct {
			Name string `json:"name" yaml:"name"`
		}),
		ResourceTemplates: []framework.ResourceTemplate{{
			// This is the generated resource the input will patch
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }}
spec:
  replicas: 1
  selector:
    app: foo
  template:
    spec:
      containers:
      - name: app
        image: example.io/team/app
`),
		}},
		MergeResources: true,
	}

	// The second resource will be treated as a patch since its metadata matches the resource
	// generated by ResourceTemplates and MergeResources is true.
	rw := kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- kind: Deployment
  apiVersion: apps/v1
  metadata: 
    name: custom
  spec:
    replicas: 6
  selector:
    app: custom
  template:
    spec:
      containers:
      - name: app
        image: example.io/team/custom
- kind: Deployment
  apiVersion: apps/v1
  metadata:
    name: mergeTest
  spec:
    replicas: 6
functionConfig:
  name: mergeTest
`)}
	if err := framework.Execute(p, &rw); err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1alpha1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: mergeTest
	//   spec:
	//     replicas: 6
	//     selector:
	//       app: foo
	//     template:
	//       spec:
	//         containers:
	//         - name: app
	//           image: example.io/team/app
	// - kind: Deployment
	//   apiVersion: apps/v1
	//   metadata:
	//     name: custom
	//   spec:
	//     replicas: 6
	//   selector:
	//     app: custom
	//   template:
	//     spec:
	//       containers:
	//       - name: app
	//         image: example.io/team/custom
	// functionConfig:
	//   name: mergeTest
}

// ExampleSelector_templatizeKinds provides an example of using a template as a selector value,
// to dynamically match resources based on the functionConfig input. It also shows how Selector
// can be used with SimpleProcessor to implement a ResourceListProcessor the filters the input.
func ExampleSelector_templatizeKinds() {
	type api struct {
		KindName string `yaml:"kindName"`
	}
	rw := &kio.ByteReadWriter{
		Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
functionConfig:
  kindName: Deployment
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
    namespace: default 
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: bar
    namespace: default
`),
	}
	config := &api{}
	p := framework.SimpleProcessor{
		Config: config,
		Filter: &framework.Selector{
			TemplateData: config,
			Kinds:        []string{"{{ .KindName }}"},
		},
	}

	err := framework.Execute(p, rw)
	if err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1beta1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: foo
	//     namespace: default
	// functionConfig:
	//   kindName: Deployment
}

// ExampleSelector_templatizeKinds provides an example of using a template as a selector value,
// to dynamically match resources based on the functionConfig input. It also shows how Selector
// can be used with SimpleProcessor to implement a ResourceListProcessor the filters the input.
func ExampleSelector_templatizeAnnotations() {
	type api struct {
		Value string `yaml:"value"`
	}
	rw := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
functionConfig:
  value: bar
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
    namespace: default
    annotations:
      key: foo
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: bar
    namespace: default
    annotations:
      key: bar
`)}
	config := &api{}
	p := framework.SimpleProcessor{
		Config: config,
		Filter: &framework.Selector{
			TemplateData: config,
			Annotations:  map[string]string{"key": "{{ .Value }}"},
		},
	}

	if err := framework.Execute(p, rw); err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1beta1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: bar
	//     namespace: default
	//     annotations:
	//       key: bar
	// functionConfig:
	//   value: bar
}

// ExampleTemplateProcessor_container_patch provides an example for using TemplateProcessor to
// patch all of the containers in the input.
func ExampleTemplateProcessor_container_patch() {
	input := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - name: foo
        image: a
      - name: bar
        image: b
---
apiVersion: v1
kind: Service
metadata:
  name: foo
spec:
  selector:
    foo: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
spec:
  template:
    spec:
      containers:
      - name: foo
        image: a
      - name: baz
        image: b
---
apiVersion: v1
kind: Service
metadata:
  name: bar
spec:
  selector:
    foo: bar
`
	p := framework.TemplateProcessor{
		PatchTemplates: []framework.PatchTemplate{
			&framework.ContainerPatchTemplate{
				Templates: framework.StringTemplates(`
env:
- name: KEY
  value: {{ .Value }}
`),
				TemplateData: struct{ Value string }{Value: "new-value"},
			}},
	}
	err := framework.Execute(p, &kio.ByteReadWriter{Reader: bytes.NewBufferString(input)})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: foo
	//         image: a
	//         env:
	//         - name: KEY
	//           value: new-value
	//       - name: bar
	//         image: b
	//         env:
	//         - name: KEY
	//           value: new-value
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	// spec:
	//   selector:
	//     foo: bar
	// ---
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: foo
	//         image: a
	//         env:
	//         - name: KEY
	//           value: new-value
	//       - name: baz
	//         image: b
	//         env:
	//         - name: KEY
	//           value: new-value
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: bar
	// spec:
	//   selector:
	//     foo: bar
}

// PatchTemplateContainersWithString patches containers matching
// a specific name.
func ExampleTemplateProcessor_container_patch_by_name() {
	input := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - name: foo
        image: a
        env:
        - name: EXISTING
          value: variable
      - name: bar
        image: b
---
apiVersion: v1
kind: Service
metadata:
  name: foo
spec:
  selector:
    foo: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
spec:
  template:
    spec:
      containers:
      - name: foo
        image: a
      - name: baz
        image: b
---
apiVersion: v1
kind: Service
metadata:
  name: bar
spec:
  selector:
    foo: bar
`
	p := framework.TemplateProcessor{
		TemplateData: struct{ Value string }{Value: "new-value"},
		PatchTemplates: []framework.PatchTemplate{
			&framework.ContainerPatchTemplate{
				// Only patch containers named "foo"
				ContainerMatcher: framework.ContainerNameMatcher("foo"),
				Templates: framework.StringTemplates(`
env:
- name: KEY
  value: {{ .Value }}
`),
			}},
	}

	err := framework.Execute(p, &kio.ByteReadWriter{Reader: bytes.NewBufferString(input)})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: foo
	//         image: a
	//         env:
	//         - name: EXISTING
	//           value: variable
	//         - name: KEY
	//           value: new-value
	//       - name: bar
	//         image: b
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: foo
	// spec:
	//   selector:
	//     foo: bar
	// ---
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: foo
	//         image: a
	//         env:
	//         - name: KEY
	//           value: new-value
	//       - name: baz
	//         image: b
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: bar
	// spec:
	//   selector:
	//     foo: bar
}

type v1alpha1JavaSpringBoot struct {
	Metadata Metadata                   `yaml:"metadata" json:"metadata"`
	Spec     v1alpha1JavaSpringBootSpec `yaml:"spec" json:"spec"`
}

type Metadata struct {
	Name string `yaml:"name" json:"name"`
}

type v1alpha1JavaSpringBootSpec struct {
	Replicas int    `yaml:"replicas" json:"replicas"`
	Domain   string `yaml:"domain" json:"domain"`
	Image    string `yaml:"image" json:"image"`
}

func (a v1alpha1JavaSpringBoot) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	filter := framework.TemplateProcessor{
		ResourceTemplates: []framework.ResourceTemplate{{
			TemplateData: &a,
			Templates: framework.StringTemplates(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Metadata.Name }}
  selector:
    app: {{ .Metadata.Name }}
spec:
  replicas: {{ .Spec.Replicas }}
  template:
    spec:
      containers:
      - name: app
        image: {{ .Spec.Image }}
        {{ if .Spec.Domain }}
        ports:
          - containerPort: 80
            name: http
        {{ end }}

{{ if .Spec.Domain }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Metadata.Name }}-svc
spec:
  selector:
    app:  {{ .Metadata.Name }}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name:  {{ .Metadata.Name }}-ingress
spec:
  tls:
  - hosts:
      - {{ .Spec.Domain }}
    secretName: secret-tls
  defaultBackend:
    service:
      name: {{ .Metadata.Name }}
      port:
        number: 80
{{ end }}
`),
		}},
	}
	return filter.Filter(items)
}

func (a *v1alpha1JavaSpringBoot) Default() error {
	if a.Spec.Replicas == 0 {
		a.Spec.Replicas = 3
	}
	return nil
}

func (a *v1alpha1JavaSpringBoot) Validate() error {
	if a.Metadata.Name == "" {
		return errors.Errorf("Name is required")
	}
	return nil
}

// ExampleVersionedAPIProcessor shows how to use the VersionedAPIProcessor and TemplateProcessor to
// build functions that implement complex multi-version APIs that require defaulting and validation.
func ExampleVersionedAPIProcessor() {
	p := &framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"JavaSpringBoot": {
			"example.com/v1alpha1": &v1alpha1JavaSpringBoot{},
		}}}

	source := &kio.ByteReadWriter{Reader: bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1beta1
kind: ResourceList
functionConfig:
  apiVersion: example.com/v1alpha1 
  kind: JavaSpringBoot
  metadata:
    name: my-app
  spec:
    image: example.docker.com/team/app:1.0
    domain: demo.example.com
`)}
	if err := framework.Execute(p, source); err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1beta1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: my-app
	//     selector:
	//       app: my-app
	//   spec:
	//     replicas: 3
	//     template:
	//       spec:
	//         containers:
	//         - name: app
	//           image: example.docker.com/team/app:1.0
	//           ports:
	//           - containerPort: 80
	//             name: http
	// - apiVersion: v1
	//   kind: Service
	//   metadata:
	//     name: my-app-svc
	//   spec:
	//     selector:
	//       app: my-app
	//     ports:
	//     - protocol: TCP
	//       port: 80
	//       targetPort: 80
	// - apiVersion: networking.k8s.io/v1
	//   kind: Ingress
	//   metadata:
	//     name: my-app-ingress
	//   spec:
	//     tls:
	//     - hosts:
	//       - demo.example.com
	//       secretName: secret-tls
	//     defaultBackend:
	//       service:
	//         name: my-app
	//         port:
	//           number: 80
	// functionConfig:
	//   apiVersion: example.com/v1alpha1
	//   kind: JavaSpringBoot
	//   metadata:
	//     name: my-app
	//   spec:
	//     image: example.docker.com/team/app:1.0
	//     domain: demo.example.com
}
