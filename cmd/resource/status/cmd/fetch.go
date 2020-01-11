// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/resource/status/generateddocs/commands"
	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// GetFetchRunner returns a command FetchRunner.
func GetFetchRunner() *FetchRunner {
	r := &FetchRunner{
		newResolverFunc: newResolver,
	}
	c := &cobra.Command{
		Use:     "fetch DIR...",
		Short:   commands.FetchShort,
		Long:    commands.FetchLong,
		Example: commands.FetchExamples,
		RunE:    r.runE,
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")

	r.Command = c
	return r
}

func FetchCommand() *cobra.Command {
	return GetFetchRunner().Command
}

// FetchRunner captures the parameters for the command and contains
// the run function.
type FetchRunner struct {
	IncludeSubpackages bool
	Command            *cobra.Command

	newResolverFunc newResolverFunc
}

func (r *FetchRunner) runE(c *cobra.Command, args []string) error {
	ctx := context.Background()

	resolver, err := r.newResolverFunc(time.Minute)
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

	// Pass in the inventory of resources to the FetchAndResolve function
	// on the resolver. It will return the status (or an error) for each
	// resource in the inventory.
	results := resolver.FetchAndResolve(ctx, captureFilter.Identifiers)

	// Create new printer that knows how to print resource statuses
	// in a table format and ask it to print the results.
	newTablePrinter(FetchStatusInfo{results}, c.OutOrStdout(), c.OutOrStderr(), false).Print()
	return nil
}

// FetchStatusInfo wraps the results from the FetchAndResolve function
// to the format expected in the TablePrinter.
type FetchStatusInfo struct {
	Results []wait.ResourceResult
}

// CurrentStatus returns the latest information known about the
// status of each of the resources. For FetchStatusInfo, the result
// is never updated, so it just returns the information provided
// by the slice of wait.ResourceResult at creation.
func (f FetchStatusInfo) CurrentStatus() StatusData {
	var resourceData []ResourceStatusData
	for _, res := range f.Results {
		rsd := ResourceStatusData{
			Identifier: res.ResourceIdentifier,
		}
		if res.Error != nil {
			rsd.Status = status.UnknownStatus
			rsd.Message = res.Error.Error()
		} else {
			rsd.Status = res.Result.Status
			rsd.Message = res.Result.Message
		}

		resourceData = append(resourceData, rsd)
	}

	return StatusData{
		AggregateStatus:  status.UnknownStatus,
		ResourceStatuses: resourceData,
	}
}
