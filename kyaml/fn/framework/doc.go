// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package framework contains a framework for writing functions in go.  The function spec
// is defined at: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
//
// Functions are executables which generate, modify, delete or validate Kubernetes resources.
// They are often used used to implement abstractions ("kind: JavaSpringBoot") and
// cross-cutting logic ("kind: SidecarInjector").
//
// Functions may be run as standalone executables or invoked as part of an orchestrated
// pipeline (e.g. kustomize).
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
//    # run the function using the
//    $ ./example-fn config.yaml patch.yaml
//
// Go implementation
//
//   // example-fn/main.go
//   func main() {
//
//     // Define the template used to generate resources
//     tc := framework.TemplateCommand{
//       Merge: true, // apply inputs as patches to the template output
//       API: &struct {
//         Key   string `json:"key" yaml:"key"`
//         Value string `json:"value" yaml:"value"`
//       }{},
//       Template: template.Must(template.New("example").Parse(`
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: foo
//     namespace: default
//     annotations:
//       {{ .Key }}: {{ .Value }}
//   `))}
//
//     // Run the command
//     if err := tc.GetCommand().Execute(); err != nil {
//       fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
//       os.Exit(1)
//     }
//  }
//
// More Examples
//
// Example function implementation using framework.Command with flag input
//
//      var value string
//      resourceList := &framework.ResourceList{}
//      cmd := framework.Command(resourceList, func() error {
//        for i := range resourceList.Items {
//          // modify the items...
//        }
//        return nil
//      })
//      cmd.Flags().StringVar(&value, "value", "", "annotation value")
//      if err := cmd.Execute(); err != nil { return err }
//
// Example function implementation using framework.ResourceList with a struct input
//
//    type Spec struct {
//      Value string `yaml:"value,omitempty"`
//    }
//    type Example struct {
//      Spec Spec `yaml:"spec,omitempty"`
//    }
//    functionConfig := &Example{}
//
//    rl := framework.ResourceList{FunctionConfig: functionConfig}
//    if err := rl.Read(); err != nil { return err	}
//
//    for i := range rl.Items {
//      // modify the items...
//    }
//    if err := rl.Write(); err != nil { return err }
//
// Architecture
//
// Functions modify a slice of resources (ResourceList.Items) which are read as input and written
// as output.  The function itself may be configured through a functionConfig
// (ResourceList.FunctionConfig).
//
// Example Function Input:
//
//    kind: ResourceList
//    items:
//    - kind: Deployment
//      ...
//    - kind: Service
//      ....
//    functionConfig:
//      kind: Example
//      spec:
//        value: foo
//
// The functionConfig may be specified declaratively and run with
//
//  config run DIR/
//
// Declarative function declaration:
//
//    kind: Example
//    metadata:
//      annotations:
//        # run the function by creating this container and providing this
//        # Example as the functionConfig
//        config.kubernetes.io/function: |
//          image: image/containing/function:impl
//    spec:
//      value: foo
//
// The framework takes care of serializing and deserializing the ResourceList.
//
// Generated ResourceList.functionConfig -- ConfigMaps
//
// Functions may also be specified imperatively and run using:
//
//   config run DIR/ --image image/containing/function:impl -- value=foo
//
// When run imperatively, a ConfigMap is generated for the functionConfig, and the command
// arguments are set as ConfigMap data entries.
//
//    kind: ConfigMap
//    data:
//      value: foo
//
// To write a function that can be run imperatively on the commandline, have it take a
// ConfigMap as its functionConfig.
//
// Mutator and Generator Functions
//
// Functions may add, delete or modify resources by modifying the ResourceList.Items slice.
//
// Validator Functions
//
// A function may emit validation results by setting the ResourceList.Result
//
// Configuring Functions
//
// Functions may be configured through a functionConfig (i.e. a client side custom resource),
// or through flags (which the framework parses from a ConfigMap provided as input).
//
// When using framework.Command, any flags registered on the cobra.Command will be parsed
// from the functionConfig input if they are defined as functionConfig.data entries.
//
// When using framework.ResourceList, any flags set on the ResourceList.Flags will be
// parsed from the functionConfig input if they are defined as functionConfig.data entries.
//
// Functions may also access environment variables set by the caller.
//
// Building a container image for the function
//
// The go program may be built into a container and run as a function.  The framework
// can be used to generate a Dockerfile to build the function container.
//
//   # create the ./Dockerfile for the container
//   $ go run ./main.go gen ./
//
//   # build the function's container
//   $ docker build . -t gcr.io/my-project/my-image:my-version
package framework
