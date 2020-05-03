// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package framework contains a framework for writing functions in go.
//
// Example
//
// Example function implementation to set an annotation on each resource.
//
//      cmd := framework.Command(nil, func(items []*yaml.RNode) ([]*yaml.RNode, error) {
//        for i := range items {
//          if err := items[i].PipeE(yaml.SetAnnotation("value", value)); err != nil {
//            return nil, err
//          }
//        }
//        return items, nil
//      })
//      cmd.Flags().StringVar(&value, "value", "", "annotation value")
//      if err := cmd.Execute(); err != nil {
//        panic(err)
//      }
//
// Architecture
//
// Functions are implemented as a go function which accept a slice of resources (items)
// and returns a modified slice of resources (items).
//
// Mutator and Generator Functions
//
// Functions may add, delete or modify resources for the returned slice.
//
// Validator Functions
//
// Functions may validate resources, returning results as go errors.  results may contain
// different items for different validation failures.
//
// Configuring Functions
//
// Functions may be configured through a functionConfig (i.e. a client side custom resource),
// or through flags (which the framework parses from a ConfigMap provided as input).
// Any flags registered on the cobra.Command will be parsed from the functionConfig input
// if they are defined as functionConfig.data entries.
//
// Functions may also access environment variables set by the caller.
//
// Function Input
//
// The framework parses the function ResourceList.items into a slice of yaml.RNodes, and
// parses the ResourceList.functionConfig into a passed in struct (optional).
//
// Building the Container
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
