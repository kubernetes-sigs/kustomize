// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework_test

import (
	"bytes"
	"fmt"

	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
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
