// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package yaml contains low-level libraries for manipulating individual Kubernetes Resource
// Configuration yaml.
//
// It exports the public pieces of "gopkg.in/yaml.v3", so can be used as a drop in replacement.
//
// This package should be used over sigs.k8s.io/yaml:
// - If retaining or modifying yaml comments, structure, formatting
// - If Resources should be round tripped without dropping fields
package yaml
