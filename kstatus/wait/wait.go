// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
)

// ResourceIdentifier defines the functions needed to identify
// a resource in a cluster. This interface is implemented by
// both unstructured.Unstructured and the standard Kubernetes types.
type ResourceIdentifier interface {
	GetName() string
	GetNamespace() string
	GetAPIVersion() string
	GetKind() string
}

// Resolver provides the functions for resolving status of a list of resources.
type Resolver struct {
	// DynamicClient is the client used to talk
	// with the cluster
	client client.Reader

	// statusComputeFunc defines which function should be used for computing
	// the status of a resource. This is available for testing purposes.
	statusComputeFunc func(u *unstructured.Unstructured) (*status.Result, error)

	// pollInterval defines the frequency with which the resolver should poll
	// the cluster for the state of resources. More frequent polling will
	// lead to more load on the cluster.
	pollInterval time.Duration
}

// NewResolver creates a new resolver with the provided client. Fetching
// and polling of resources will be done using the provided client.
func NewResolver(client client.Reader, pollInterval time.Duration) *Resolver {
	return &Resolver{
		client:            client,
		statusComputeFunc: status.Compute,
		pollInterval:      pollInterval,
	}
}

// ResourceResult is the status result for a given resource. It provides
// information about the resource if the request was successful and an
// error if something went wrong.
type ResourceResult struct {
	Result *status.Result

	Resource ResourceIdentifier

	Error error
}

// FetchAndResolve returns the status for a list of resources. It will return
// the status for each of them individually. The slice of ResourceIdentifiers will
// only be used to get the information needed to fetch the updated state of
// the resources from the cluster.
func (r *Resolver) FetchAndResolve(ctx context.Context, resources []ResourceIdentifier) []ResourceResult {
	var results []ResourceResult

	for _, resource := range resources {
		u, err := r.fetchResource(ctx, resource)
		if err != nil {
			if k8serrors.IsNotFound(errors.Cause(err)) {
				results = append(results, ResourceResult{
					Resource: resource,
					Result: &status.Result{
						Status:  status.CurrentStatus,
						Message: "Resource does not exist",
					},
				})
			} else {
				results = append(results, ResourceResult{
					Result: &status.Result{
						Status:  status.UnknownStatus,
						Message: fmt.Sprintf("Error fetching resource from cluster: %v", err),
					},
					Resource: resource,
					Error:    err,
				})
			}
			continue
		}
		res, err := r.statusComputeFunc(u)
		results = append(results, ResourceResult{
			Result:   res,
			Resource: resource,
			Error:    err,
		})
	}

	return results
}

// Event is returned through the channel returned after a call
// to WaitForStatus. It contains an update to either an individual
// resource or to the aggregate status for the set of resources.
type Event struct {
	// Type defines which type of event this is.
	Type EventType

	// AggregateStatus is the aggregated status for all the provided resources.
	AggregateStatus status.Status

	// EventResource is information about the event to which this event pertains.
	// This is only populated for ResourceUpdate events.
	EventResource *EventResource
}

type EventType string

const (
	// The status/message for a resource has changed. This also means the
	// aggregate status might have changed.
	ResourceUpdate EventType = "ResourceUpdate"

	// All resources have reached the current status.
	Completed EventType = "Completed"

	// The wait was stopped before all resources could reach the
	// Current status.
	Aborted EventType = "Aborted"
)

// EventResource contains information about the resource for which
// a specific Event pertains.
type EventResource struct {
	// Identifier contains information that identifies which resource
	// this information is about.
	Identifier ResourceIdentifier

	// Status is the latest status for the given resource.
	Status status.Status

	// Message is more details about the status.
	Message string

	// Error is set if there was a problem identifying the status
	// of the resource. For example, if polling the cluster for information
	// about the resource failed.
	Error error
}

// WaitForStatus polls all the provided resources until all of them has
// reached the Current status. Updates the channel as resources change their status and
// when the wait is either completed or aborted.
func (r *Resolver) WaitForStatus(ctx context.Context, resources []ResourceIdentifier) <-chan Event {
	eventChan := make(chan Event)

	go func() {
		ticker := time.NewTicker(r.pollInterval)

		defer func() {
			ticker.Stop()
			// Make sure the channel is closed so consumers can detect that
			// we have completed.
			close(eventChan)
		}()

		// No need to wait if we have no resources. We consider
		// this a situation where the status is Current.
		if len(resources) == 0 {
			eventChan <- Event{
				Type:            Completed,
				AggregateStatus: status.CurrentStatus,
				EventResource:   nil,
			}
			return
		}

		// Initiate a new waitStatus object to keep track of the
		// resources while polling the state.
		waitState := newWaitState(resources, r.statusComputeFunc)

		// Check all resources immediately. If the aggregate status is already
		// Current, we can exit immediately.
		if r.checkAllResources(ctx, waitState, eventChan) {
			return
		}

		// Loop until either all resources have reached the Current status
		// or until the wait is cancelled through the context. In both cases
		// we will break out of the loop by returning from the function.
		for {
			select {
			case <-ctx.Done():
				// The context has been cancelled, so report the most recent
				// aggregate status, report it through the channel and then
				// break out of the loop (which will close the channel).
				eventChan <- Event{
					Type:            Aborted,
					AggregateStatus: waitState.AggregateStatus(),
				}
				return
			case <-ticker.C:
				// Every time the ticker fires, we check the status of all
				// resources. If the aggregate status has reached Current, checkAllResources
				// will return true. If so, we just return.
				if r.checkAllResources(ctx, waitState, eventChan) {
					return
				}
			}
		}
	}()

	return eventChan
}

// checkAllResources fetches all resources from the cluster,
// checks if their status has changed and send an event for each resource
// with a new status. In each event, we also include the latest aggregate
// status. Finally, if the aggregate status becomes Current, send a final
// Completed type event. If the aggregate status has become Current, this function
// will return true to signal that it is done.
func (r *Resolver) checkAllResources(ctx context.Context, waitState *waitState, eventChan chan Event) bool {
	for id := range waitState.ResourceWaitStates {
		// Make sure we have a local copy since we are passing
		// pointers to this variable as parameters to functions
		identifier := id
		u, err := r.fetchResource(ctx, &identifier)
		eventResource, updateObserved := waitState.ResourceObserved(&identifier, u, err)
		// Find the aggregate status based on the new state for this resource.
		aggStatus := waitState.AggregateStatus()
		// We want events for changes in status for each resource, so send
		// an event for this resource before checking if the aggregate status
		// has become Current.
		if updateObserved {
			eventChan <- Event{
				Type:            ResourceUpdate,
				AggregateStatus: aggStatus,
				EventResource:   &eventResource,
			}
		}
		// If aggregate status is Current, we are done!
		if aggStatus == status.CurrentStatus {
			eventChan <- Event{
				Type:            Completed,
				AggregateStatus: status.CurrentStatus,
			}
			return true
		}
	}
	return false
}

// fetchResource gets the resource given by the identifier from the cluster
// through the client available in the Resolver. It returns the resource
// as an Unstructured.
func (r *Resolver) fetchResource(ctx context.Context, identifier ResourceIdentifier) (*unstructured.Unstructured, error) {
	key := types.NamespacedName{Name: identifier.GetName(), Namespace: identifier.GetNamespace()}
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": identifier.GetAPIVersion(),
			"kind":       identifier.GetKind(),
		},
	}
	err := r.client.Get(ctx, key, u)
	//return u, err
	if err != nil {
		return nil, errors.Wrap(err, "error fetching resource from cluster")
	}
	return u, nil
}
