// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/resource/status/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// GetEventsRunner returns a command EventsRunner.
func GetEventsRunner() *EventsRunner {
	r := &EventsRunner{
		newResolverFunc: newResolver,
	}
	c := &cobra.Command{
		Use:     "events DIR...",
		Short:   commands.EventsShort,
		Long:    commands.EventsLong,
		Example: commands.EventsExamples,
		RunE:    r.runE,
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().DurationVar(&r.Interval, "interval", 2*time.Second,
		"check every n seconds.")
	c.Flags().DurationVar(&r.Timeout, "timeout", 60*time.Second,
		"give up after n seconds.")

	r.Command = c
	return r
}

func EventsCommand() *cobra.Command {
	return GetEventsRunner().Command
}

// EventsRunner captures the parameters for the command
// and contains the run function.
type EventsRunner struct {
	IncludeSubpackages bool
	Interval           time.Duration
	Timeout            time.Duration
	Command            *cobra.Command

	newResolverFunc newResolverFunc
}

func (r *EventsRunner) runE(c *cobra.Command, args []string) error {
	ctx := context.Background()

	resolver, err := r.newResolverFunc(r.Interval)
	if err != nil {
		return errors.Wrap(err, "error creating resolver")
	}

	// Set up a CaptureIdentifierFilter and run all inputs through the
	// filter with the pipeline to capture the inventory of resources
	// which we are interested in.
	captureFilter := &CaptureIdentifiersFilter{}
	filters := []kio.Filter{captureFilter}

	var inputs []kio.Reader
	for _, a := range args {
		inputs = append(inputs, kio.LocalPackageReader{
			PackagePath:        a,
			IncludeSubpackages: r.IncludeSubpackages,
		})
	}
	if len(inputs) == 0 {
		inputs = append(inputs, &kio.ByteReader{Reader: c.InOrStdin()})
	}

	err = kio.Pipeline{
		Inputs:  inputs,
		Filters: filters,
	}.Execute()
	if err != nil {
		return errors.Wrap(err, "error reading manifests")
	}

	// Create a new printer that knows how to print updates about
	// resourdes and their aggregate status in the events format.
	printer := newEventPrinter(c.OutOrStdout(), c.OutOrStderr())

	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	resChannel := resolver.WaitForStatus(ctx, captureFilter.Identifiers)

	// Print events until the channel is closed. This will happen
	// either because all resources has reached the Current status
	// or it has timed out.
	for msg := range resChannel {
		printer.printEvent(msg)
	}
	return nil
}
