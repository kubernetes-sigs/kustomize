// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// +build ignore

package main

import (
	"github.com/kubeflow/kubeflow/bootstrap/v3/pkg/apis/apps/kfdef/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
)

// plugin loads the KfDefChart CRD scheme into kustomize
type plugin struct {
	ldr ifc.Loader
	rf  *resmap.Factory
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, _ []byte) (err error) {
	p.ldr = ldr
	p.rf = rf

	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	return v1beta1.SchemeBuilder.AddToScheme(scheme.Scheme)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	return nil
}
