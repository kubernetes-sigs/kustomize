// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main contains an example using the the framework.
// The example annotates all resources in the input with the value provided as a flag.
//
// To execute the function, run:
//
//   $ cat input/cm.yaml | go run ./main.go --value=foo
//
// To generate the Dockerfile for the function image run:
//
//   $ go run ./main.go gen ./
package main
