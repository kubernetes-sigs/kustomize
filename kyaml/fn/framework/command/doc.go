// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package command contains a builder for creating cobra.Commands based on configuration functions
// written using the kyaml function framework. The commands this package generates can be used as
// standalone executables or as part of a configuration management pipeline that complies with the
// Configuration Functions Specification (e.g. Kustomize generators or transformers):
// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
//
// Example standalone usage
//
// Function template input:
//
//    # config.yaml -- this is the input to the template
//    apiVersion: example.com/v1alpha1
//    kind: Example
//    Key: a
//    Value: b
//
// Additional function inputs:
//
//    # patch.yaml -- this will be applied as a patch
//    apiVersion: apps/v1
//    kind: Deployment
//    metadata:
//      name: foo
//      namespace: default
//      annotations:
//        patch-key: patch-value
//
// Manually run the function:
//
//    # build the function
//    $ go build example-fn/
//
//    # run the function
//    $ ./example-fn config.yaml patch.yaml
//
// Go implementation
//
//   // example-fn/main.go
//	func main() {
//		// Define the template used to generate resources
//		p := framework.TemplateProcessor{
//			MergeResources: true, // apply inputs as patches to the template output
//			TemplateData: new(struct {
//				Key   string `json:"key" yaml:"key"`
//				Value string `json:"value" yaml:"value"`
//			}),
//			ResourceTemplates: []framework.ResourceTemplate{{
//				Templates: framework.StringTemplates(`
//	  apiVersion: apps/v1
//	  kind: Deployment
//	  metadata:
//		name: foo
//		namespace: default
//		annotations:
//		  {{ .Key }}: {{ .Value }}
//	  `)}},
//		}
//
//		// Run the command
//		if err := command.Build(p, command.StandaloneEnabled, true).Execute(); err != nil {
//			fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
//			os.Exit(1)
//		}
//	}
//
// Example function implementation using command.Build with flag input
//
//	func main() {
//		var value string
//		fn := func(rl *framework.ResourceList) error {
//			for i := range rl.Items {
//				// set the annotation on each resource item
//				if err := rl.Items[i].PipeE(yaml.SetAnnotation("value", value)); err != nil {
//					return err
//				}
//			}
//			return nil
//		}
//		cmd := command.Build(framework.ResourceListProcessorFunc(fn), command.StandaloneEnabled, false)
//		cmd.Flags().StringVar(&value, "value", "", "annotation value")
//
//		if err := cmd.Execute(); err != nil {
//			fmt.Println(err)
//			os.Exit(1)
//		}
//	}
package command
