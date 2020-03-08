// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package starlark contains a kio.Filter which can be applied to resources to transform
// them through starlark program.
//
// The resources are provided to the program through the global variable "resourceList".
// "resourceList" is a dictionary containing an "items" field with a list of resources.
// Changes to "resourceList" made by the starlark program will be reflected in the Filter output.
//
// After being run through the starlark program, the filter will copy the comments from the input
// resources to restore them after they are dropped due to the serialization.
//
// The Filter will also format the output so that output has the preferred field ordering
// rather than an alphabetical field ordering.
//
// The resourceList variable adheres to the kustomize function spec as specified by:
// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
//
// All items in the resourceList are resources represented as starlark dictionaries/
// The items in the resourceList respect the io spec specified by:
// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/config-io.md
//
// The starlark language spec can be found here:
// https://github.com/google/starlark-go/blob/master/doc/spec.md
package starlark
