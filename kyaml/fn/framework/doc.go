// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package framework contains a framework for writing functions in go.  The function spec
// is defined at: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
//
// Examples
//
// Example function implementation using framework.ResourceList with functionConfig
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
// Example function implementation using framework.Command with flags
//
//      var value string
//      cmd := framework.Command(nil, func(items []*yaml.RNode) ([]*yaml.RNode, error) {
//        for i := range items {
//          // modify the items...
//        }
//        return items, nil
//      })
//      cmd.Flags().StringVar(&value, "value", "", "annotation value")
//      if err := cmd.Execute(); err != nil { return err }
//
// Architecture
//
// Functions modify a slice of resources (ResourceList.items) which are read as input and written
// as output.  The function itself may be configured through a functionConfig
// (ResourceList.functionConfig).
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
//          image: image/containing/fuction:impl
//    spec:
//      value: foo
//
// The framework takes care of serializing and deserializing the ResourceList.
//
// Generated ResourceList.functionConfig -- ConfigMaps
//
// Functions may also be specified imperatively and run using:
//
//   config run DIR/ --image image/containing/fuction:impl -- value=foo
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
// Functions may add, delete or modify resources by modifying the items slice.
// When using framework.Command this is done through returning the new items slice.
// When using framework.ResourceList this is done through modifying ResourceList.Items in place.
//
// Validator Functions
//
// A function may validate resources by providing a Result.
// When using framework.Command this is done through returning a framework.Result as an error.
// WHen using framework.ResourceList this is done through setting ResourceList.Result.
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
// The go program must be built into a container to be run as a function.  The framework
// can be used to generate a Dockerfile to build the function container.
//
//   # create the ./Dockerfile for the container
//   $ go run ./main.go gen ./
//
//   # build the function's container
//   $ docker build . -t gcr.io/my-project/my-image:my-version
package framework
