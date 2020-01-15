// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/resource/status/generateddocs/commands"
	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// GetWaitRunner return a command WaitRunner.
func GetWaitRunner() *WaitRunner {
	r := &WaitRunner{
		newResolverFunc: newResolver,
	}
	c := &cobra.Command{
		Use:     "wait DIR...",
		Short:   commands.WaitShort,
		Long:    commands.WaitLong,
		Example: commands.WaitExamples,
		RunE:    r.runE,
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

func WaitCommand() *cobra.Command {
	return GetWaitRunner().Command
}

// WaitRunner captures the parameters for the command and contains
// the run function.
type WaitRunner struct {
	IncludeSubpackages bool
	Interval           time.Duration
	Timeout            time.Duration
	Command            *cobra.Command

	newResolverFunc newResolverFunc
}

// runE implements the logic of the command and will call the Wait command in the wait
// package, use a ResourceStatusCollector to capture the events from the channel, and the
// TablePrinter to display the information.
func (r *WaitRunner) runE(c *cobra.Command, args []string) error {
	ctx := context.Background()

	resolver, err := r.newResolverFunc(r.Interval)
	if err != nil {
		return errors.Wrap(err, "errors creating resolver")
	}

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

	collector := newResourceStatusCollector(captureFilter.Identifiers)

	stop := make(chan struct{})
	printer := newTablePrinter(CollectorStatusInfo{collector}, c.OutOrStdout(), c.OutOrStderr(), true)
	printFinished := printer.PrintUntil(stop, 1*time.Second)

	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	resChannel := resolver.WaitForStatus(ctx, captureFilter.Identifiers)

	for msg := range resChannel {
		switch msg.Type {
		case wait.ResourceUpdate:
			collector.updateResourceStatus(msg)
		case wait.Aborted:
			collector.updateAggregateStatus(msg.AggregateStatus)
		case wait.Completed:
			collector.updateAggregateStatus(msg.AggregateStatus)
		}
	}
	close(stop)
	<-printFinished // Wait for printer to finish work.
	return nil
}

// ResourceStatusCollector captures the latest state seen for all resources
// based on the events from the Wait channel. This is used by the TablePrinter
// to display status for all resources.
type ResourceStatusCollector struct {
	mux sync.RWMutex

	AggregateStatus  status.Status
	ResourceStatuses []*ResourceStatus
}

// updateResourceStatus takes the given event and update the status info
// in the ResourceStatusCollector.
func (r *ResourceStatusCollector) updateResourceStatus(msg wait.Event) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.AggregateStatus = msg.AggregateStatus
	eventResource := msg.EventResource
	for _, resourceState := range r.ResourceStatuses {
		if resourceState.Identifier.GroupKind == eventResource.ResourceIdentifier.GroupKind &&
			resourceState.Identifier.Namespace == eventResource.ResourceIdentifier.Namespace &&
			resourceState.Identifier.Name == eventResource.ResourceIdentifier.Name {
			resourceState.Status = eventResource.Status
			resourceState.Message = eventResource.Message
		}
	}
}

// updateAggregateStatus sets the aggregate status of the ResourceStatusCollector to the
// given value.
func (r *ResourceStatusCollector) updateAggregateStatus(aggregateStatus status.Status) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.AggregateStatus = aggregateStatus
}

// ResourceStatus contains the status information for a single resource.
type ResourceStatus struct {
	Identifier wait.ResourceIdentifier
	Status     status.Status
	Message    string
}

// newResourceStatusCollector creates a new ResourceStatusCollector with the given
// resources and sets the status for all of them to Unknown.
func newResourceStatusCollector(identifiers []wait.ResourceIdentifier) *ResourceStatusCollector {
	var statuses []*ResourceStatus

	for _, id := range identifiers {
		statuses = append(statuses, &ResourceStatus{
			Identifier: id,
			Status:     status.UnknownStatus,
			Message:    "",
		})
	}

	return &ResourceStatusCollector{
		AggregateStatus:  status.UnknownStatus,
		ResourceStatuses: statuses,
	}
}

// CollectorStatusInfo is a wrapper around the ResourceStatusCollector
// to make it adhere to the interface of the TableWriter.
type CollectorStatusInfo struct {
	Collector *ResourceStatusCollector
}

// CurrentStatus implements the interface for the TableWriter and
// returns a copy of the current status of the resources in the
// ResourceStatusCollector. This is done to make sure the TableWriter
// does not have to deal with synchronization when accessing the data.
func (f CollectorStatusInfo) CurrentStatus() StatusData {
	f.Collector.mux.RLock()
	defer f.Collector.mux.RUnlock()

	var resourceData []ResourceStatusData
	for _, res := range f.Collector.ResourceStatuses {
		resourceData = append(resourceData, ResourceStatusData{
			Identifier: res.Identifier,
			Status:     res.Status,
			Message:    res.Message,
		})
	}

	return StatusData{
		AggregateStatus:  f.Collector.AggregateStatus,
		ResourceStatuses: resourceData,
	}
}
