// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package krmfile provides functionality for working with Krmfiles.
//
// Example Krmfile
//
//    apiVersion: config.k8s.io/v1alpha1
//    kind: Krmfile
//    openAPI:
//      definitions:
//        io.k8s.cli.setters.replicas:
//          x-k8s-cli:
//            setter:
//              name: replicas
//              value: "3"
//              setBy: me
//          description: "hello world"
package krmfile
