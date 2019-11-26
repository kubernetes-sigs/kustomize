package wait

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/kstatus/status"
)

// resourceKey is a minimal implementation of
// the ResourceIdentifier interface.
type resourceKey struct {
	name       string
	namespace  string
	apiVersion string
	kind       string
}

// GetName returns the name of the resource.
func (r *resourceKey) GetName() string {
	return r.name
}

// GetNamespace returns the namespace of the resource.
func (r *resourceKey) GetNamespace() string {
	return r.namespace
}

// GetAPIVersion returns the API version of the resource.
func (r *resourceKey) GetAPIVersion() string {
	return r.apiVersion
}

// GetKind returns the Kind of the resource.
func (r *resourceKey) GetKind() string {
	return r.kind
}

// waitState keeps the state about the resources and their last
// observed state. This is used to determine any changes in state
// so events can be sent when needed.
type waitState struct {
	// ResourceWaitStates contains wait state for each of the resources.
	ResourceWaitStates map[resourceKey]*resourceWaitState

	// statusComputeFunc defines the function used to compute the state of
	// a single resource. This is available for testing purposes.
	statusComputeFunc func(u *unstructured.Unstructured) (*status.Result, error)
}

// resourceWaitState contains state information about an individual resource.
type resourceWaitState struct {
	FirstSeenGeneration *int64
	HasBeenCurrent      bool
	Observed            bool

	LastEvent *EventResource
}

// newWaitState creates a new waitState object and initializes it with the
// provided slice of resources and the provided statusComputeFunc.
func newWaitState(resources []ResourceIdentifier, statusComputeFunc func(u *unstructured.Unstructured) (*status.Result, error)) *waitState {
	resourceWaitStates := make(map[resourceKey]*resourceWaitState)

	for _, r := range resources {
		identifier := resourceKey{
			apiVersion: r.GetAPIVersion(),
			kind:       r.GetKind(),
			name:       r.GetName(),
			namespace:  r.GetNamespace(),
		}
		resourceWaitStates[identifier] = &resourceWaitState{}
	}

	return &waitState{
		ResourceWaitStates: resourceWaitStates,
		statusComputeFunc:  statusComputeFunc,
	}
}

// AggregateStatus computes the aggregate status for all the resources.
// TODO: Ideally we would like this to be pluggable for different strategies.
func (w *waitState) AggregateStatus() status.Status {
	allCurrent := true
	for _, rws := range w.ResourceWaitStates {
		if !rws.Observed {
			return status.UnknownStatus
		}
		if !rws.HasBeenCurrent {
			allCurrent = false
		}
	}
	if allCurrent {
		return status.CurrentStatus
	}
	return status.InProgressStatus
}

// ResourceObserved notifies the waitState that we have new state for
// a resource. This also accepts an error in case fetching the resource
// from a cluster failed. It returns an EventResource object that contains
// information about the observed resource, including the identifier and
// the latest status for the resource. The function also returns a bool value
// that will be true if the status of the observed resource has changed
// since the previous observation and false it not. This is used to determine
// whether a new event should be sent based on this observation.
func (w *waitState) ResourceObserved(id ResourceIdentifier, resource *unstructured.Unstructured, err error) (EventResource, bool) {
	identifier := resourceKey{
		name:       id.GetName(),
		namespace:  id.GetNamespace(),
		apiVersion: id.GetAPIVersion(),
		kind:       id.GetKind(),
	}

	// Check for nil is not needed here as the id passed in comes
	// from iterating over the keys of the map.
	rws := w.ResourceWaitStates[identifier]

	eventResource := w.getEventResource(identifier, resource, err)
	// If the new eventResource is identical to the previous one, we return
	// with the last return value indicating this is not a new event.
	if rws.LastEvent != nil && reflect.DeepEqual(eventResource, *rws.LastEvent) {
		return eventResource, false
	}
	rws.LastEvent = &eventResource
	return eventResource, true
}

// getEventResource creates a new EventResource for the resource identified by
// the provided resourceKey. The EventResource contains information about the
// latest status for the given resource, so it computes status for the resource
// as well as check for deletion.
func (w *waitState) getEventResource(identifier resourceKey, resource *unstructured.Unstructured, err error) EventResource {
	// Get the resourceWaitState for this resource. It contains information
	// of the previous observed statuses. We don't need to check for nil here
	// as the identifier comes from iterating over the keys of the
	// ResourceWaitState map.
	r := w.ResourceWaitStates[identifier]

	// If fetching the resource from the cluster failed, we don't really
	// know anything about the status of the resource, so simply
	// report the status as Unknown.
	if err != nil && !k8serrors.IsNotFound(errors.Cause(err)) {
		return EventResource{
			Identifier: &identifier,
			Status:     status.UnknownStatus,
			Message:    fmt.Sprintf("Error: %s", err),
			Error:      err,
		}
	}

	// If we get here, we have successfully fetched the resource from
	// the cluster, or discovered that it doesn't exist.
	r.Observed = true

	// We treat a non-existent resource as Current. This is to properly
	// handle deletion scenarios.
	if k8serrors.IsNotFound(errors.Cause(err)) {
		r.HasBeenCurrent = true
		return EventResource{
			Identifier: &identifier,
			Status:     status.CurrentStatus,
			Message:    fmt.Sprintf("Resource has been deleted"),
		}
	}

	// We want to capture the first seen generation of the resource. This
	// allows us to discover if a resource is updated while we are waiting
	// for it to become Current.
	if r.FirstSeenGeneration != nil {
		gen := resource.GetGeneration()
		r.FirstSeenGeneration = &gen
	}

	if resource.GetDeletionTimestamp() != nil {
		return EventResource{
			Identifier: &identifier,
			Status:     status.TerminatingStatus,
			Message:    fmt.Sprintf("Resource is terminating"),
		}
	}

	statusResult, err := w.statusComputeFunc(resource)
	// If we can't compute status for the resource, we report the status
	// as Unknown.
	if err != nil {
		return EventResource{
			Identifier: &identifier,
			Status:     status.UnknownStatus,
			Message:    fmt.Sprintf("Error: %s", err),
			Error:      err,
		}
	}

	// We record whether a resource has ever been Current. This makes
	// sure we can report a set of resources as being Current if all
	// of them has reached the Current status at some point, but not
	// necessarily at the same time.
	if statusResult.Status == status.CurrentStatus {
		r.HasBeenCurrent = true
	}

	return EventResource{
		Identifier: &identifier,
		Status:     statusResult.Status,
		Message:    statusResult.Message,
	}
}
