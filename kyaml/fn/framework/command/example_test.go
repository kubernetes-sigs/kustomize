// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package command_test

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const service = "Service"

// ExampleBuild_modify implements a function that sets an annotation on each resource.
// The annotation value is configured via ResourceList.FunctionConfig.
func ExampleBuild_modify() {
	// create a struct matching the structure of ResourceList.FunctionConfig to hold its data
	var config struct {
		Data map[string]string `yaml:"data"`
	}
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range items {
			// set the annotation on each resource item
			err := items[i].PipeE(yaml.SetAnnotation("value", config.Data["value"]))
			if err != nil {
				return nil, err
			}
		}
		return items, nil
	}
	p := framework.SimpleProcessor{Filter: kio.FilterFunc(fn), Config: &config}
	cmd := command.Build(p, command.StandaloneDisabled, false)

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
	//   - apiVersion: apps/v1
	//     kind: Deployment
	//     metadata:
	//       name: foo
	//       annotations:
	//         value: 'baz'
	//   - apiVersion: v1
	//     kind: Service
	//     metadata:
	//       name: foo
	//       annotations:
	//         value: 'baz'
	// functionConfig:
	//   apiVersion: v1
	//   kind: ConfigMap
	//   data:
	//     value: baz
}

// ExampleBuild_generateReplace generates a resource from a FunctionConfig.
// If the resource already exists, it replaces the resource with a new copy.
func ExampleBuild_generateReplace() {
	// function API definition which will be parsed from the ResourceList.FunctionConfig
	// read from stdin
	type Spec struct {
		Name string `yaml:"name,omitempty"`
	}
	type ExampleServiceGenerator struct {
		Spec Spec `yaml:"spec,omitempty"`
	}
	functionConfig := &ExampleServiceGenerator{}

	// function implementation -- generate a Service resource
	p := &framework.SimpleProcessor{
		Config: functionConfig,
		Filter: kio.FilterFunc(func(items []*yaml.RNode) ([]*yaml.RNode, error) {
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

			// generate the resource
			n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
 name: %s
`, functionConfig.Spec.Name))
			if err != nil {
				return nil, err
			}
			newNodes = append(newNodes, n)
			return newNodes, nil
		}),
	}
	cmd := command.Build(p, command.StandaloneDisabled, false)

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
	//   - apiVersion: apps/v1
	//     kind: Deployment
	//     metadata:
	//       name: foo
	//   - apiVersion: v1
	//     kind: Service
	//     metadata:
	//       name: bar
	// functionConfig:
	//   apiVersion: example.com/v1alpha1
	//   kind: ExampleServiceGenerator
	//   spec:
	//     name: bar
}

// ExampleBuild_generateUpdate generates a resource, updating the previously generated
// copy rather than replacing it.
//
// Note: This will keep manual edits to the previously generated copy.
func ExampleBuild_generateUpdate() {
	// function API definition which will be parsed from the ResourceList.FunctionConfig
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
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		var found bool
		for i := range items {
			meta, err := items[i].GetMeta()
			if err != nil {
				return nil, err
			}

			// something we already generated, reconcile it to make sure it matches what
			// is specified by the FunctionConfig
			if meta.Name == functionConfig.Spec.Name &&
				meta.Kind == service &&
				meta.APIVersion == "v1" {
				// set some values
				for k, v := range functionConfig.Spec.Annotations {
					err := items[i].PipeE(yaml.SetAnnotation(k, v))
					if err != nil {
						return nil, err
					}
				}
				found = true
				break
			}
		}
		if found {
			return items, nil
		}

		// generate the resource if not found
		n, err := yaml.Parse(fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
 name: %s
`, functionConfig.Spec.Name))
		if err != nil {
			return nil, err
		}
		for k, v := range functionConfig.Spec.Annotations {
			err := n.PipeE(yaml.SetAnnotation(k, v))
			if err != nil {
				return nil, err
			}
		}
		items = append(items, n)
		return items, nil
	}

	p := &framework.SimpleProcessor{Config: functionConfig, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)

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
	//   - apiVersion: apps/v1
	//     kind: Deployment
	//     metadata:
	//       name: foo
	//   - apiVersion: v1
	//     kind: Service
	//     metadata:
	//       name: bar
	//       annotations:
	//         a: 'b'
	// functionConfig:
	//   apiVersion: example.com/v1alpha1
	//   kind: ExampleServiceGenerator
	//   spec:
	//     name: bar
	//     annotations:
	//       a: b
}

// ExampleBuild_validate validates that all Deployment resources have the replicas field set.
// If any Deployments do not contain spec.replicas, then the function will return results
// which will be set on ResourceList.results
func ExampleBuild_validate() {
	fn := func(rl *framework.ResourceList) error {
		// validation results
		var validationResults []framework.ResultItem

		// validate that each Deployment resource has spec.replicas set
		for i := range rl.Items {
			// only check Deployment resources
			meta, err := rl.Items[i].GetMeta()
			if err != nil {
				return err
			}
			if meta.Kind != "Deployment" {
				continue
			}

			// lookup replicas field
			r, err := rl.Items[i].Pipe(yaml.Lookup("spec", "replicas"))
			if err != nil {
				return err
			}

			// check replicas not specified
			if r != nil {
				continue
			}
			validationResults = append(validationResults, framework.ResultItem{
				Severity: framework.Error,
				Message:  "field is required",
				ResourceRef: yaml.ResourceIdentifier{
					TypeMeta: meta.TypeMeta,
					NameMeta: meta.ObjectMeta.NameMeta,
				},
				Field: framework.Field{
					Path:           "spec.replicas",
					SuggestedValue: "1",
				},
			})
		}

		if len(validationResults) > 0 {
			rl.Result = &framework.Result{
				Name:  "replicas-validator",
				Items: validationResults,
			}
		}

		return rl.Result
	}

	cmd := command.Build(framework.ResourceListProcessorFunc(fn), command.StandaloneDisabled, true)
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
	//   - apiVersion: apps/v1
	//     kind: Deployment
	//     metadata:
	//       name: foo
	// results:
	//   name: replicas-validator
	//   items:
	//     - message: field is required
	//       severity: error
	//       resourceRef:
	//         apiVersion: apps/v1
	//         kind: Deployment
	//         metadata:
	//           name: foo
	//       field:
	//         path: spec.replicas
	//         suggestedValue: "1"
}
