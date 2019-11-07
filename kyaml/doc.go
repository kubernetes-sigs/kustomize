// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package kyaml contains libraries for reading and writing Kubernetes Resource configuration
// as yaml.
//
// Resources
//
// Individual Resources are manipulated using the yaml package.
//  import (
//      "sigs.k8s.io/kustomize/kyaml/yaml"
//  )
//
// Collections of Resources
//
// Collections of Resources are manipulated using the kio package.
//  import (
//      "sigs.k8s.io/kustomize/kyaml/kio"
//  )
package kyaml
