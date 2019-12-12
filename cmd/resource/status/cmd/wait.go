package cmd

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func GetWaitRunner() *WaitRunner {
	r := &WaitRunner{}
	c := &cobra.Command{
		Use:   "wait",
		Short: "Wait",
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

func WaitCommand() *cobra.Command {
	return GetWaitRunner().Command
}

type WaitRunner struct {
	IncludeSubpackages bool
	Interval           time.Duration
	Timeout            time.Duration
	Command            *cobra.Command
}

func (r *WaitRunner) runE(c *cobra.Command, args []string) error {
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

type ResourceStatusCollector struct {
	mux sync.RWMutex

	AggregateStatus  status.Status
	ResourceStatuses []*ResourceStatus
}

func (r *ResourceStatusCollector) updateResourceStatus(msg wait.Event) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.AggregateStatus = msg.AggregateStatus
	eventResource := msg.EventResource
	for _, resourceState := range r.ResourceStatuses {
		if resourceState.Identifier.GetAPIVersion() == eventResource.Identifier.GetAPIVersion() &&
			resourceState.Identifier.GetKind() == eventResource.Identifier.GetKind() &&
			resourceState.Identifier.GetNamespace() == eventResource.Identifier.GetNamespace() &&
			resourceState.Identifier.GetName() == eventResource.Identifier.GetName() {
			resourceState.Status = eventResource.Status
			resourceState.Message = eventResource.Message
		}
	}
}

func (r *ResourceStatusCollector) updateAggregateStatus(aggregateStatus status.Status) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.AggregateStatus = aggregateStatus
}

type ResourceStatus struct {
	Identifier wait.ResourceIdentifier
	Status     status.Status
	Message    string
}

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

type CollectorStatusInfo struct {
	Collector *ResourceStatusCollector
}

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
