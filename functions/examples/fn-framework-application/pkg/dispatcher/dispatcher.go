// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package dispatcher

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/functions/examples/fn-framework-application/pkg/exampleapp/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func New() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(processKnownAPIGroups)
}

func NewCommand() *cobra.Command {
	return command.Build(New(), command.StandaloneEnabled, false)
}

func processKnownAPIGroups(rl *framework.ResourceList) error {
	p := framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"ExampleApp": map[string]kio.Filter{
			"platform.example.com/v1alpha1": &v1alpha1.ExampleApp{},
		},
	}}
	if err := p.Process(rl); err != nil {
		return errors.Wrap(err)
	}
	var err error
	// FormatFilter sorts the fields in a deterministic order, which makes the output
	// more suitable for review in version control.
	rl.Items, err = filters.FormatFilter{UseSchema: true}.Filter(rl.Items)
	if err != nil {
		return errors.WrapPrefixf(err, "formatting output")
	}
	return nil
}
