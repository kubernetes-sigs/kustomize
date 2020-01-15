// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

type createClientFunc func() (client.Reader, error)

// createClient returns a client for talking to a Kubernetes cluster. The client
// is from controller-runtime.
func createClient() (client.Reader, error) {
	config := ctrl.GetConfigOrDie()
	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, err
	}
	return client.New(config, client.Options{Scheme: scheme, Mapper: mapper})
}

func newClientFunc(c client.Reader) func() (client.Reader, error) {
	return func() (client.Reader, error) {
		return c, nil
	}
}

// CaptureIdentifiersFilter implements the Filter interface in the kio package. It
// captures the identifiers for all resources passed through the pipeline.
type CaptureIdentifiersFilter struct {
	Identifiers []wait.ResourceIdentifier
}

var _ kio.Filter = &CaptureIdentifiersFilter{}

func (f *CaptureIdentifiersFilter) Filter(slice []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range slice {
		meta, err := slice[i].GetMeta()
		if err != nil {
			return nil, err
		}
		id := meta.GetIdentifier()
		if IsValidKubernetesResource(&id) {
			f.Identifiers = append(f.Identifiers, &id)
		}
	}
	return slice, nil
}

func IsValidKubernetesResource(id *yaml.ResourceIdentifier) bool {
	return id != nil && id.GetKind() != "" && id.GetAPIVersion() != "" && id.GetName() != ""
}
