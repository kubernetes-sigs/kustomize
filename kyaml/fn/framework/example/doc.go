// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main contains an example using the the framework.
// The example annotates all resources in the input with the value provided as a flag,
// and adds all resources in the templates/ directory to the list.
//
// To execute the function, run:
//
//   $ cat testdata/basic/input.yaml | go run ./main.go --value=foo
//
// Alternatively, you can provide the value via a config file instead of a flag:
//
//   $ go run ./main.go testdata/basic/config.yaml testdata/basic/input.yaml
//
// To generate the Dockerfile for the function image run:
//
//   $ go run ./main.go gen ./
package main
