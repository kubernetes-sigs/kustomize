// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const service = "Service"

// ExampleResourceList_modify implements a function that sets an annotation on each resource.
// The annotation value is configured via a flag value parsed from ResourceList.functionConfig.data
func ExampleResourceList_modify() {
	// for testing purposes only -- normally read from stdin when Executing
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
# functionConfig is parsed into flags by framework.Command
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  data:
    value: baz
`)

	// configure the annotation value using a flag parsed from
	// ResourceList.functionConfig.data.value
	fs := pflag.NewFlagSet("tests", pflag.ContinueOnError)
	value := fs.String("value", "", "annotation value")
	rl := framework.ResourceList{
		Flags:  fs,
		Reader: input, // for testing only
	}
	if err := rl.Read(); err != nil {
		panic(err)
	}
	for i := range rl.Items {
		// set the annotation on each resource item
		if err := rl.Items[i].PipeE(yaml.SetAnnotation("value", *value)); err != nil {
			panic(err)
		}
	}
	if err := rl.Write(); err != nil {
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

// ExampleCommand_modify implements a function that sets an annotation on each resource.
// The annotation value is configured via a flag value parsed from
// ResourceList.functionConfig.data
func ExampleCommand_modify() {
	// configure the annotation value using a flag parsed from
	// ResourceList.functionConfig.data.value
	resourceList := framework.ResourceList{}
	var value string
	cmd := framework.Command(&resourceList, func() error {
		for i := range resourceList.Items {
			// set the annotation on each resource item
			err := resourceList.Items[i].PipeE(yaml.SetAnnotation("value", value))
			if err != nil {
				return err
			}
		}
		return nil
	})
	cmd.Flags().StringVar(&value, "value", "", "annotation value")

	// for testing purposes only -- normally read from stdin when Executing
	cmd.SetIn(bytes.NewBufferString(`
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
# functionConfig is parsed into flags by framework.Command
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  data:
    value: baz
`))
	// run the command
	if err := cmd.Execute(); err != nil {
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

// ExampleCommand_generateReplace generates a resource from a functionConfig.
// If the resource already exist s, it replaces the resource with a new copy.
func ExampleCommand_generateReplace() {
	// function API definition which will be parsed from the ResourceList.functionConfig
	// read from stdin
	type Spec struct {
		Name string `yaml:"name,omitempty"`
	}
	type ExampleServiceGenerator struct {
		Spec Spec `yaml:"spec,omitempty"`
	}
	functionConfig := &ExampleServiceGenerator{}

	// function implementation -- generate a Service resource
	resourceList := &framework.ResourceList{FunctionConfig: functionConfig}
	cmd := framework.Command(resourceList, func() error {
		var newNodes []*yaml.RNode
		for i := range resourceList.Items {
			meta, err := resourceList.Items[i].GetMeta()
			if err != nil {
				return err
			}

			// something we already generated, remove it from the list so we regenerate it
			if meta.Name == functionConfig.Spec.Name &&
				meta.Kind == service &&
				meta.APIVersion == "v1" {
				continue
			}
			newNodes = append(newNodes, resourceList.Items[i])
		}

		// generate the resource
		n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
`, functionConfig.Spec.Name))
		if err != nil {
			return err
		}
		newNodes = append(newNodes, n)
		resourceList.Items = newNodes
		return nil
	})

	// for testing purposes only -- normally read from stdin when Executing
	cmd.SetIn(bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
# items are provided as nodes
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
# functionConfig is parsed into flags by framework.Command
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: ExampleServiceGenerator
  spec:
    name: bar
`))

	// run the command
	if err := cmd.Execute(); err != nil {
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

// ExampleResourceList_generateReplace generates a resource from a functionConfig.
// If the resource already exist s, it replaces the resource with a new copy.
func ExampleResourceList_generateReplace() {
	input := bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
# items are provided as nodes
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
# functionConfig is parsed into flags by framework.Command
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: ExampleServiceGenerator
  spec:
    name: bar
`)

	// function API definition which will be parsed from the ResourceList.functionConfig
	// read from stdin
	type Spec struct {
		Name string `yaml:"name,omitempty"`
	}
	type ExampleServiceGenerator struct {
		Spec Spec `yaml:"spec,omitempty"`
	}
	functionConfig := &ExampleServiceGenerator{}

	rl := framework.ResourceList{
		FunctionConfig: functionConfig,
		Reader:         input, // for testing only
	}
	if err := rl.Read(); err != nil {
		panic(err)
	}

	// remove the last generated resource
	var newNodes []*yaml.RNode
	for i := range rl.Items {
		meta, err := rl.Items[i].GetMeta()
		if err != nil {
			panic(err)
		}
		// something we already generated, remove it from the list so we regenerate it
		if meta.Name == functionConfig.Spec.Name &&
			meta.Kind == service &&
			meta.APIVersion == "v1" {
			continue
		}
		newNodes = append(newNodes, rl.Items[i])
	}
	rl.Items = newNodes

	// generate the resource again
	n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
`, functionConfig.Spec.Name))
	if err != nil {
		panic(err)
	}
	rl.Items = append(rl.Items, n)

	if err := rl.Write(); err != nil {
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

// ExampleCommand_generateUpdate generates a resource, updating the previously generated
// copy rather than replacing it.
//
// Note: This will keep manual edits to the previously generated copy.
func ExampleCommand_generateUpdate() {
	// function API definition which will be parsed from the ResourceList.functionConfig
	// read from stdin
	type Spec struct {
		Name        string            `yaml:"name,omitempty"`
		Annotations map[string]string `yaml:"annotations,omitempty"`
	}
	type ExampleServiceGenerator struct {
		Spec Spec `yaml:"spec,omitempty"`
	}
	functionConfig := &ExampleServiceGenerator{}

	// function implementation -- generate or update a Service resource
	resourceList := &framework.ResourceList{FunctionConfig: functionConfig}
	cmd := framework.Command(resourceList, func() error {
		var found bool
		for i := range resourceList.Items {
			meta, err := resourceList.Items[i].GetMeta()
			if err != nil {
				return err
			}

			// something we already generated, reconcile it to make sure it matches what
			// is specified by the functionConfig
			if meta.Name == functionConfig.Spec.Name &&
				meta.Kind == service &&
				meta.APIVersion == "v1" {
				// set some values
				for k, v := range functionConfig.Spec.Annotations {
					err := resourceList.Items[i].PipeE(yaml.SetAnnotation(k, v))
					if err != nil {
						return err
					}
				}
				found = true
				break
			}
		}
		if found {
			return nil
		}

		// generate the resource if not found
		n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
`, functionConfig.Spec.Name))
		if err != nil {
			return err
		}
		for k, v := range functionConfig.Spec.Annotations {
			err := n.PipeE(yaml.SetAnnotation(k, v))
			if err != nil {
				return err
			}
		}
		resourceList.Items = append(resourceList.Items, n)

		return nil
	})

	// for testing purposes only -- normally read from stdin when Executing
	cmd.SetIn(bytes.NewBufferString(`
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
    name: bar
# functionConfig is parsed into flags by framework.Command
functionConfig:
  apiVersion: example.com/v1alpha1
  kind: ExampleServiceGenerator
  spec:
    name: bar
    annotations:
      a: b
`))

	// run the command
	if err := cmd.Execute(); err != nil {
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
	//     annotations:
	//       a: 'b'
	// functionConfig:
	//   apiVersion: example.com/v1alpha1
	//   kind: ExampleServiceGenerator
	//   spec:
	//     name: bar
	//     annotations:
	//       a: b
}

// ExampleCommand_validate validates that all Deployment resources have the replicas field set.
// If any Deployments do not contain spec.replicas, then the function will return results
// which will be set on ResourceList.results
func ExampleCommand_validate() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		// validation results
		var validationResults []framework.Item

		// validate that each Deployment resource has spec.replicas set
		for i := range resourceList.Items {
			// only check Deployment resources
			meta, err := resourceList.Items[i].GetMeta()
			if err != nil {
				return err
			}
			if meta.Kind != "Deployment" {
				continue
			}

			// lookup replicas field
			r, err := resourceList.Items[i].Pipe(yaml.Lookup("spec", "replicas"))
			if err != nil {
				return err
			}

			// check replicas not specified
			if r != nil {
				continue
			}
			validationResults = append(validationResults, framework.Item{
				Severity:    framework.Error,
				Message:     "missing replicas",
				ResourceRef: meta,
				Field: framework.Field{
					Path:           "spec.field",
					SuggestedValue: "1",
				},
			})
		}

		if len(validationResults) > 0 {
			resourceList.Result = &framework.Result{
				Name:  "replicas-validator",
				Items: validationResults,
			}
		}

		return resourceList.Result
	})

	// for testing purposes only -- normally read from stdin when Executing
	cmd.SetIn(bytes.NewBufferString(`
apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
# items are provided as nodes
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: foo
`))

	// run the command
	if err := cmd.Execute(); err != nil {
		// normally exit 1 here
	}

	// Output:
	// apiVersion: config.kubernetes.io/v1alpha1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: foo
	// results:
	//   name: replicas-validator
	//   items:
	//   - message: missing replicas
	//     severity: error
	//     resourceRef:
	//       apiVersion: apps/v1
	//       kind: Deployment
	//       metadata:
	//         name: foo
	//     field:
	//       path: spec.field
	//       suggestedValue: "1"
}

// ExampleTemplateCommand provides an example for using the TemplateCommand
func ExampleTemplateCommand() {
	// create the template
	cmd := framework.TemplateCommand{
		// Template input
		API: &struct {
			Key   string `json:"key" yaml:"key"`
			Value string `json:"value" yaml:"value"`
		}{},
		// Template
		Template: template.Must(template.New("example").Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  annotations:
    {{ .Key }}: {{ .Value }}
`)),
	}.GetCommand()

	cmd.SetArgs([]string{filepath.Join("testdata", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
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

// ExampleTemplateCommand_files provides an example for using the TemplateCommand
func ExampleTemplateCommand_files() {
	// create the template
	cmd := framework.TemplateCommand{
		// Template input
		API: &struct {
			Key   string `json:"key" yaml:"key"`
			Value string `json:"value" yaml:"value"`
		}{},
		// Template
		TemplatesFiles: []string{filepath.Join("testdata", "templatefiles", "deployment.template")},
	}.GetCommand()

	cmd.SetArgs([]string{filepath.Join("testdata", "templatefiles", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
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

// ExampleTemplateCommand_preprocess provides an example for using the TemplateCommand
// with PreProcess to configure the template based on the input resources observed.
func ExampleTemplateCommand_preprocess() {
	config := &struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
		Short bool
	}{}

	// create the template
	cmd := framework.TemplateCommand{
		// Template input
		API: config,
		PreProcess: func(rl *framework.ResourceList) error {
			config.Short = len(rl.Items) < 3
			return nil
		},
		// Template
		Template: template.Must(template.New("example").Parse(`
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
`)),
	}.GetCommand()

	cmd.SetArgs([]string{filepath.Join("testdata", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
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

// ExampleTemplateCommand_postprocess provides an example for using the TemplateCommand
// with PostProcess to modify the results.
func ExampleTemplateCommand_postprocess() {
	config := &struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
	}{}

	// create the template
	cmd := framework.TemplateCommand{
		// Template input
		API: config,
		// Template
		Template: template.Must(template.New("example").Parse(`
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
`)),
		PostProcess: func(rl *framework.ResourceList) error {
			// trim the first resources
			rl.Items = rl.Items[1:]
			return nil
		},
	}.GetCommand()

	cmd.SetArgs([]string{filepath.Join("testdata", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
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

// ExampleTemplateCommand_patch provides an example for using the TemplateCommand to
// create a function which patches resources.
func ExampleTemplateCommand_patch() {
	// patch the foo resource only
	s := framework.Selector{Names: []string{"foo"}}

	cmd := framework.TemplateCommand{
		API: &struct {
			Key   string `json:"key" yaml:"key"`
			Value string `json:"value" yaml:"value"`
		}{},
		Template: template.Must(template.New("example").Parse(`
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
`)),
		// PatchTemplates are applied to BOTH ResourceList input resources AND templated resources
		PatchTemplates: []framework.PatchTemplate{{
			Selector: &s,
			Template: template.Must(template.New("test").Parse(`
metadata:
  annotations:
    patched: 'true'
`)),
		}},
	}.GetCommand()

	cmd.SetArgs([]string{filepath.Join("testdata", "template", "config.yaml")})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
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

func ExampleSelector_templatizeKinds() {
	type api struct {
		KindName string `yaml:"kindName"`
	}
	rl := &framework.ResourceList{
		FunctionConfig: &api{KindName: "Deployment"},
		Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: bar
  namespace: default
`),
		Writer: os.Stdout,
	}
	if err := rl.Read(); err != nil {
		panic(err)
	}

	var err error
	s := &framework.Selector{
		TemplatizeValues: true,
		Kinds:            []string{"{{ .KindName }}"},
	}
	rl.Items, err = s.GetMatches(rl)
	if err != nil {
		panic(err)
	}

	if err := rl.Write(); err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: foo
	//   namespace: default
	//   annotations:
	//     config.kubernetes.io/index: '0'
}

func ExampleSelector_templatizeAnnotations() {
	type api struct {
		Value string `yaml:"vaue"`
	}
	rl := &framework.ResourceList{
		FunctionConfig: &api{Value: "bar"},
		Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
  annotations:
    key: foo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  namespace: default
  annotations:
    key: bar
`),
		Writer: os.Stdout,
	}
	if err := rl.Read(); err != nil {
		panic(err)
	}

	var err error
	s := &framework.Selector{
		TemplatizeValues: true,
		Annotations:      map[string]string{"key": "{{ .Value }}"},
	}
	rl.Items, err = s.GetMatches(rl)
	if err != nil {
		panic(err)
	}

	if err := rl.Write(); err != nil {
		panic(err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: bar
	//   namespace: default
	//   annotations:
	//     key: bar
	//     config.kubernetes.io/index: '1'
}

// ExamplePatchContainersWithString patches all containers.
func ExamplePatchContainersWithString() {
	resources, err := kio.ParseAll(`
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
`)
	if err != nil {
		log.Fatal(err)
	}

	input := struct{ Value string }{Value: "new-value"}
	err = framework.PatchContainersWithString(resources, `
env:
  KEY: {{ .Value }}
`, input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(kio.StringAll(resources))

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
	//           KEY: new-value
	//       - name: bar
	//         image: b
	//         env:
	//           KEY: new-value
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
	//           KEY: new-value
	//       - name: baz
	//         image: b
	//         env:
	//           KEY: new-value
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: bar
	// spec:
	//   selector:
	//     foo: bar
	//  <nil>
}

// PatchTemplateContainersWithString patches containers matching
// a specific name.
func ExamplePatchContainersWithString_names() {
	resources, err := kio.ParseAll(`
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
`)
	if err != nil {
		log.Fatal(err)
	}

	input := struct{ Value string }{Value: "new-value"}
	err = framework.PatchContainersWithString(resources, `
env:
  KEY: {{ .Value }}
`, input, "foo")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(kio.StringAll(resources))

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
	//           KEY: new-value
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
	//           KEY: new-value
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
	//  <nil>
}
