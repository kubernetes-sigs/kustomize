package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func GetEventsRunner() *EventsRunner {
	r := &EventsRunner{}
	c := &cobra.Command{
		Use:   "events",
		Short: "Events",
		RunE:  r.runE,
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")
	c.Flags().DurationVar(&r.Interval, "interval", 2*time.Second,
		"check every n seconds. Default is every 2 seconds.")
	c.Flags().DurationVar(&r.Timeout, "timeout", 60*time.Second,
		"give up after n seconds. Default is 60 seconds.")

	r.Command = c
	return r
}

func EventsCommand() *cobra.Command {
	return GetEventsRunner().Command
}

type EventsRunner struct {
	IncludeSubpackages bool
	Interval           time.Duration
	Timeout            time.Duration
	Command            *cobra.Command
}

func (r *EventsRunner) runE(c *cobra.Command, args []string) error {
	ctx := context.Background()
	client, err := getClient()
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}

	resolver := wait.NewResolver(client, r.Interval)

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

	printer := newEventPrinter(c.OutOrStdout(), c.OutOrStderr())

	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	resChannel := resolver.WaitForStatus(ctx, captureFilter.Identifiers)

	for msg := range resChannel {
		printer.printEvent(msg)
	}
	return nil
}
