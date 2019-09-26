// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"istio.io/istio/pilot/pkg/config/kube/crd"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
)

// plugin loads the Network CRD scheme into kustomize
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

	// SchemeGroupVersion is group version used to register these objects
	apiVersion := schema.GroupVersion{Group: "networking.istio.io", Version: "v1alpha3"}
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			for _, kind := range crd.KnownTypes {
				scheme.AddKnownTypes(apiVersion, kind.Object, kind.Collection)
			}
			meta_v1.AddToGroupVersion(scheme, apiVersion)
			return nil
		})

	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	return schemeBuilder.AddToScheme(scheme.Scheme)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	return nil
}
