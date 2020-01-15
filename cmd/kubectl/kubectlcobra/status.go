// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

type StatusOptions struct {
	factory   util.Factory
	ioStreams genericclioptions.IOStreams

	wait    bool
	period  time.Duration
	timeout time.Duration
}

func newStatusOptions(factory util.Factory, ioStreams genericclioptions.IOStreams) *StatusOptions {
	return &StatusOptions{
		factory:   factory,
		ioStreams: ioStreams,

		wait:    false,
		period:  2 * time.Second,
		timeout: 1 * time.Minute,
	}
}

func (s *StatusOptions) AddFlags(c *cobra.Command) {
	c.Flags().BoolVar(&s.wait, "status", s.wait, "Wait for all applied resources to reach the Current status.")
	c.Flags().DurationVar(&s.period, "status-period", s.period, "Polling period for resource statuses.")
	c.Flags().DurationVar(&s.timeout, "status-timeout", s.timeout, "Timeout threshold for waiting for all resources to reach the Current status.")
}

func (s *StatusOptions) waitForStatus(infos []*resource.Info) error {
	mapper, err := getRESTMapper(s.factory)
	if err != nil {
		return err
	}

	c, err := getClient(s.factory, mapper)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resolver := wait.NewResolver(c, mapper, s.period)
	ch := resolver.WaitForStatus(ctx, infosToResourceIdentifiers(infos))

	for msg := range ch {
		switch msg.Type {
		case wait.ResourceUpdate:
			id := msg.EventResource.ResourceIdentifier
			gk := id.GroupKind
			fmt.Fprintf(s.ioStreams.Out, "%s/%s is %s: %s\n", strings.ToLower(gk.String()), id.Name, msg.EventResource.Status.String(), msg.EventResource.Message)
		case wait.Completed:
			fmt.Fprint(s.ioStreams.Out, "all resources has reached the Current status\n")
		case wait.Aborted:
			fmt.Fprintf(s.ioStreams.Out, "resources failed to the reached Current status after %s\n", s.timeout.String())
		}
	}
	return nil
}

func infosToResourceIdentifiers(infos []*resource.Info) []wait.ResourceIdentifier {
	var resources []wait.ResourceIdentifier
	for _, info := range infos {
		u := info.Object.(*unstructured.Unstructured)
		resources = append(resources, wait.ResourceIdentifier{
			GroupKind: u.GroupVersionKind().GroupKind(),
			Namespace: u.GetNamespace(),
			Name:      u.GetName(),
		})
	}
	return resources
}

func getRESTMapper(f util.Factory) (meta.RESTMapper, error) {
	return f.ToRESTMapper()
}

func getClient(f util.Factory, mapper meta.RESTMapper) (client.Reader, error) {
	config, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
}
