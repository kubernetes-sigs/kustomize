// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

type newResolverFunc func(pollInterval time.Duration) (*wait.Resolver, error)

// newResolver returns a new resolver that can resolve status for resources based
// on polling the cluster.
func newResolver(pollInterval time.Duration) (*wait.Resolver, error) {
	config := ctrl.GetConfigOrDie()
	mapper, err := apiutil.NewDiscoveryRESTMapper(config)
	if err != nil {
		return nil, err
	}

	c, err := client.New(config, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, err
	}

	return wait.NewResolver(c, mapper, pollInterval), nil
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
		// TODO(mortent): Update kyaml library
		id := meta.GetIdentifier()
		gv, err := schema.ParseGroupVersion(id.APIVersion)
		if err != nil {
			return nil, err
		}
		if IsValidKubernetesResource(id) {
			f.Identifiers = append(f.Identifiers, wait.ResourceIdentifier{
				Name:      id.Name,
				Namespace: id.Namespace,
				GroupKind: schema.GroupKind{
					Group: gv.Group,
					Kind:  id.Kind,
				},
			})
		}
	}
	return slice, nil
}

func IsValidKubernetesResource(id yaml.ResourceIdentifier) bool {
	return id.GetKind() != "" && id.GetAPIVersion() != "" && id.GetName() != ""
}
