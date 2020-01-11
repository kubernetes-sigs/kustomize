// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
)

const (
	defaultNamespace = "default"
)

// ResourceIdentifier defines the functions needed to identify
// a resource in a cluster. This interface is implemented by
// both unstructured.Unstructured and the standard Kubernetes types.
type KubernetesObject interface {
	GetName() string
	GetNamespace() string
	GroupVersionKind() schema.GroupVersionKind
}

// ResourceIdentifier contains the information needed to uniquely
// identify a resource in a cluster.
type ResourceIdentifier struct {
	Name      string
	Namespace string
	GroupKind schema.GroupKind
}

// Equals compares two ResourceIdentifiers and returns true if they
// refer to the same resource. Special handling is needed for namespace
// since an empty namespace for a namespace-scoped resource is defaulted
// to the "default" namespace.
func (r ResourceIdentifier) Equals(other ResourceIdentifier) bool {
	isSameNamespace := r.Namespace == other.Namespace ||
		(r.Namespace == "" && other.Namespace == defaultNamespace) ||
		(r.Namespace == defaultNamespace && other.Namespace == "")
	return r.GroupKind == other.GroupKind &&
		r.Name == other.Name &&
		isSameNamespace
}

// Resolver provides the functions for resolving status of a list of resources.
type Resolver struct {
	// client is the client used to talk
	// with the cluster. It uses the Reader interface
	// from controller-runtime.
	client client.Reader

	// mapper is the RESTMapper needed to look up mappings
	// for resource types.
	mapper meta.RESTMapper

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
func NewResolver(client client.Reader, mapper meta.RESTMapper, pollInterval time.Duration) *Resolver {
	return &Resolver{
		client:            client,
		mapper:            mapper,
		statusComputeFunc: status.Compute,
		pollInterval:      pollInterval,
	}
}

// ResourceResult is the status result for a given resource. It provides
// information about the resource if the request was successful and an
// error if something went wrong.
type ResourceResult struct {
	Result *status.Result

	ResourceIdentifier ResourceIdentifier

	Error error
}

// FetchAndResolveObjects returns the status for a list of kubernetes objects. These can be provided
// either as Unstructured resources or the specific resource types. It will return the status for each
// of them individually. The provided resources will only be used to get the information needed to
// fetch the updated state of the resources from the cluster.
func (r *Resolver) FetchAndResolveObjects(ctx context.Context, objects []KubernetesObject) []ResourceResult {
	resourceIds := resourceIdentifiersFromObjects(objects)
	return r.FetchAndResolve(ctx, resourceIds)
}

// FetchAndResolve returns the status for a list of ResourceIdentifiers. It will return
// the status for each of them individually.
func (r *Resolver) FetchAndResolve(ctx context.Context, resourceIDs []ResourceIdentifier) []ResourceResult {
	var results []ResourceResult

	for _, resourceID := range resourceIDs {
		u, err := r.fetchResource(ctx, resourceID)
		if err != nil {
			if k8serrors.IsNotFound(errors.Cause(err)) {
				results = append(results, ResourceResult{
					ResourceIdentifier: resourceID,
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
					ResourceIdentifier: resourceID,
					Error:              err,
				})
			}
			continue
		}
		res, err := r.statusComputeFunc(u)
		results = append(results, ResourceResult{
			Result:             res,
			ResourceIdentifier: resourceID,
			Error:              err,
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
	ResourceIdentifier ResourceIdentifier

	// Status is the latest status for the given resource.
	Status status.Status

	// Message is more details about the status.
	Message string

	// Error is set if there was a problem identifying the status
	// of the resource. For example, if polling the cluster for information
	// about the resource failed.
	Error error
}

// WaitForStatus polls all the provided resources until all of them have reached the Current
// status or the timeout specified through the context is reached. Updates on the status
// of individual resources and the aggregate status is provided through the Event channel.
func (r *Resolver) WaitForStatusOfObjects(ctx context.Context, objects []KubernetesObject) <-chan Event {
	resourceIds := resourceIdentifiersFromObjects(objects)
	return r.WaitForStatus(ctx, resourceIds)
}

// WaitForStatus polls all the resources references by the provided ResourceIdentifiers until
// all of them have reached the Current status or the timeout specified through the context is
// reached. Updates on the status of individual resources and the aggregate status is provided
// through the Event channel.
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
	for resourceID := range waitState.ResourceWaitStates {
		// Make sure we have a local copy since we are passing
		// pointers to this variable as parameters to functions
		u, err := r.fetchResource(ctx, resourceID)
		eventResource, updateObserved := waitState.ResourceObserved(resourceID, u, err)
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
	// We need to look up the preferred version for the GroupKind and
	// whether the resource type is cluster scoped. We look this
	// up with the RESTMapper.
	mapping, err := r.mapper.RESTMapping(identifier.GroupKind)
	if err != nil {
		return nil, err
	}

	// Resources might not have the namespace set, which means we need to set
	// it to `default` if the resource is namespace scoped.
	namespace := identifier.Namespace
	if namespace == "" && mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		namespace = defaultNamespace
	}

	key := types.NamespacedName{Name: identifier.Name, Namespace: namespace}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(mapping.GroupVersionKind)
	err = r.client.Get(ctx, key, u)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching resource from cluster")
	}
	return u, nil
}
