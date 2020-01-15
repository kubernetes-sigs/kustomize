package wait

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/kustomize/kstatus/status"
)

const (
	testTimeout      = 1 * time.Minute
	testPollInterval = 1 * time.Second
)

func TestFetchAndResolve(t *testing.T) {
	type result struct {
		status status.Status
		error  bool
	}

	testCases := map[string]struct {
		resources       []runtime.Object
		mapperGVKs      []schema.GroupVersionKind
		expectedResults []result
	}{
		"no resources": {
			resources:       []runtime.Object{},
			expectedResults: []result{},
		},
		"single resource": {
			resources: []runtime.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
						Name:       "myDeployment",
						Namespace:  "default",
					},
				},
			},
			mapperGVKs: []schema.GroupVersionKind{
				appsv1.SchemeGroupVersion.WithKind("Deployment"),
			},
			expectedResults: []result{
				{
					status: status.InProgressStatus,
					error:  false,
				},
			},
		},
		"multiple resources": {
			resources: []runtime.Object{
				&appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "StatefulSet",
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
						Name:       "myStatefulSet",
						Namespace:  "default",
					},
					Spec: appsv1.StatefulSetSpec{
						UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
							Type: appsv1.OnDeleteStatefulSetStrategyType,
						},
					},
					Status: appsv1.StatefulSetStatus{
						ObservedGeneration: 1,
					},
				},
				&corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
						Name:       "mySecret",
						Namespace:  "default",
					},
				},
			},
			mapperGVKs: []schema.GroupVersionKind{
				appsv1.SchemeGroupVersion.WithKind("StatefulSet"),
				corev1.SchemeGroupVersion.WithKind("Secret"),
			},
			expectedResults: []result{
				{
					status: status.CurrentStatus,
					error:  false,
				},
				{
					status: status.CurrentStatus,
					error:  false,
				},
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, tc.resources...)

			resolver := NewResolver(fakeClient, newRESTMapper(tc.mapperGVKs...), testPollInterval)
			resolver.statusComputeFunc = status.Compute

			var identifiers []ResourceIdentifier
			for _, resource := range tc.resources {
				gvk := resource.GetObjectKind().GroupVersionKind()
				r := resource.(metav1.Object)
				identifiers = append(identifiers, ResourceIdentifier{
					Name:      r.GetName(),
					Namespace: r.GetNamespace(),
					GroupKind: gvk.GroupKind(),
				})
			}

			results := resolver.FetchAndResolve(context.TODO(), identifiers)
			for i, res := range results {
				id := identifiers[i]
				expectedRes := tc.expectedResults[i]
				rid := fmt.Sprintf("%s/%s", id.Namespace, id.Name)
				if expectedRes.error {
					if res.Error == nil {
						t.Errorf("expected error for resource %s, but didn't get one", rid)
					}
					continue
				}

				if res.Error != nil {
					t.Errorf("didn't expected error for resource %s, but got %v", rid, res.Error)
				}

				if got, want := res.Result.Status, expectedRes.status; got != want {
					t.Errorf("expected status %s for resources %s, but got %s", want, rid, got)
				}
			}
		})
	}
}

func TestFetchAndResolveUnknownResource(t *testing.T) {
	fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)
	resolver := NewResolver(fakeClient, newRESTMapper(appsv1.SchemeGroupVersion.WithKind("Deployment")), testPollInterval)
	results := resolver.FetchAndResolve(context.TODO(), []ResourceIdentifier{
		{
			GroupKind: schema.GroupKind{
				Group: "apps",
				Kind:  "Deployment",
			},
			Name:      "myDeployment",
			Namespace: "default",
		},
	})

	if want, got := 1, len(results); want != got {
		t.Errorf("expected %d results, but got %d", want, got)
	}

	res := results[0]

	if want, got := status.CurrentStatus, res.Result.Status; got != want {
		t.Errorf("expected status %s, but got %s", want, got)
	}

	if res.Error != nil {
		t.Errorf("expected no error, but got %v", res.Error)
	}
}

func TestFetchAndResolveWithFetchError(t *testing.T) {
	expectedError := errors.New("failed to fetch resource")
	resolver := NewResolver(
		&fakeReader{
			Err: expectedError,
		},
		newRESTMapper(appsv1.SchemeGroupVersion.WithKind("Deployment")),
		testPollInterval,
	)
	results := resolver.FetchAndResolve(context.TODO(), []ResourceIdentifier{
		{
			GroupKind: schema.GroupKind{
				Group: "apps",
				Kind:  "Deployment",
			},
			Name:      "myDeployment",
			Namespace: "default",
		},
	})

	if want, got := 1, len(results); want != got {
		t.Errorf("expected %d results, but got %d", want, got)
	}

	res := results[0]

	if want, got := status.UnknownStatus, res.Result.Status; got != want {
		t.Errorf("expected status %s, but got %s", want, got)
	}

	if want, got := expectedError, errors.Cause(res.Error); got != want {
		t.Errorf("expected error %v, but got %v", want, got)
	}
}

func TestFetchAndResolveComputeStatusError(t *testing.T) {
	expectedError := errors.New("this is a test")
	resource := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
			Name:       "myDeployment",
			Namespace:  "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, resource)
	resolver := NewResolver(fakeClient, newRESTMapper(appsv1.SchemeGroupVersion.WithKind("Deployment")), testPollInterval)

	resolver.statusComputeFunc = func(u *unstructured.Unstructured) (*status.Result, error) {
		return &status.Result{
			Status:  status.UnknownStatus,
			Message: "Got an error",
		}, expectedError
	}
	results := resolver.FetchAndResolve(context.TODO(), []ResourceIdentifier{
		{
			GroupKind: schema.GroupKind{
				Group: resource.GroupVersionKind().Group,
				Kind:  resource.Kind,
			},
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
	})

	if want, got := 1, len(results); want != got {
		t.Errorf("expected %d results, but got %d", want, got)
	}

	res := results[0]
	if want, got := expectedError, res.Error; got != want {
		t.Errorf("expected error %v, but got %v", want, got)
	}

	if want, got := status.UnknownStatus, res.Result.Status; got != want {
		t.Errorf("expected status %s, but got %s", want, got)
	}
}

type fakeReader struct {
	Called int
	Err    error
}

func (f *fakeReader) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	f.Called += 1
	return f.Err
}

func (f *fakeReader) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return errors.New("list not used")
}

func TestWaitForStatus(t *testing.T) {
	testCases := map[string]struct {
		resources                 map[runtime.Object][]*status.Result
		expectedResourceStatuses  map[runtime.Object][]status.Status
		expectedAggregateStatuses []status.Status
	}{
		"no resources": {
			resources:                map[runtime.Object][]*status.Result{},
			expectedResourceStatuses: map[runtime.Object][]status.Status{},
			expectedAggregateStatuses: []status.Status{
				status.CurrentStatus,
			},
		},
		"single resource": {
			resources: map[runtime.Object][]*status.Result{
				deploymentResource: {
					{
						Status:  status.InProgressStatus,
						Message: "FirstInProgress",
					},
					{
						Status:  status.InProgressStatus,
						Message: "SecondInProgress",
					},
					{
						Status:  status.CurrentStatus,
						Message: "CurrentProgress",
					},
				},
			},
			expectedResourceStatuses: map[runtime.Object][]status.Status{
				deploymentResource: {
					status.InProgressStatus,
					status.InProgressStatus,
					status.CurrentStatus,
				},
			},
			expectedAggregateStatuses: []status.Status{
				status.InProgressStatus,
				status.InProgressStatus,
				status.CurrentStatus,
				status.CurrentStatus,
			},
		},
		"multiple resource": {
			resources: map[runtime.Object][]*status.Result{
				statefulSetResource: {
					{
						Status:  status.InProgressStatus,
						Message: "FirstUnknown",
					},
					{
						Status:  status.InProgressStatus,
						Message: "SecondInProgress",
					},
					{
						Status:  status.CurrentStatus,
						Message: "CurrentProgress",
					},
				},
				serviceResource: {
					{
						Status:  status.CurrentStatus,
						Message: "CurrentImmediately",
					},
				},
			},
			expectedResourceStatuses: map[runtime.Object][]status.Status{
				statefulSetResource: {
					status.InProgressStatus,
					status.InProgressStatus,
					status.CurrentStatus,
				},
				serviceResource: {
					status.CurrentStatus,
				},
			},
			expectedAggregateStatuses: []status.Status{
				status.UnknownStatus,
				status.InProgressStatus,
				status.InProgressStatus,
				status.CurrentStatus,
				status.CurrentStatus,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			var objs []runtime.Object
			statusResults := make(map[ResourceIdentifier][]*status.Result)
			var identifiers []ResourceIdentifier

			for obj, statuses := range tc.resources {
				objs = append(objs, obj)
				identifier := resourceIdentifierFromRuntimeObject(obj)
				identifiers = append(identifiers, identifier)
				statusResults[identifier] = statuses
			}

			statusComputer := statusComputer{
				results:           statusResults,
				resourceCallCount: make(map[ResourceIdentifier]int),
			}

			resolver := &Resolver{
				client: fake.NewFakeClientWithScheme(scheme.Scheme, objs...),
				mapper: newRESTMapper(
					appsv1.SchemeGroupVersion.WithKind("Deployment"),
					appsv1.SchemeGroupVersion.WithKind("StatefulSet"),
					corev1.SchemeGroupVersion.WithKind("Service"),
				),
				statusComputeFunc: statusComputer.Compute,
				pollInterval:      testPollInterval,
			}

			eventChan := resolver.WaitForStatus(context.TODO(), identifiers)

			var events []Event
			timer := time.NewTimer(testTimeout)
		loop:
			for {
				select {
				case event, ok := <-eventChan:
					if !ok {
						break loop
					}
					events = append(events, event)
				case <-timer.C:
					t.Fatalf("timeout waiting for resources to reach current status")
				}
			}

			var aggregateStatuses []status.Status
			resourceStatuses := make(map[ResourceIdentifier][]status.Status)
			for _, e := range events {
				aggregateStatuses = append(aggregateStatuses, e.AggregateStatus)
				if e.EventResource != nil {
					identifier := e.EventResource.ResourceIdentifier
					resourceStatuses[identifier] = append(resourceStatuses[identifier], e.EventResource.Status)
				}
			}

			for resource, expectedStatuses := range tc.expectedResourceStatuses {
				identifier := resourceIdentifierFromRuntimeObject(resource)
				actualStatuses := resourceStatuses[identifier]
				if !reflect.DeepEqual(expectedStatuses, actualStatuses) {
					t.Errorf("expected statuses %v for resource %s/%s, but got %v", expectedStatuses, identifier.Namespace, identifier.Name, actualStatuses)
				}
			}

			if !reflect.DeepEqual(tc.expectedAggregateStatuses, aggregateStatuses) {
				t.Errorf("expected aggregate statuses %v, but got %v", tc.expectedAggregateStatuses, aggregateStatuses)
			}
		})
	}
}

func TestWaitForStatusDeletedResources(t *testing.T) {
	statusComputer := statusComputer{
		results:           make(map[ResourceIdentifier][]*status.Result),
		resourceCallCount: make(map[ResourceIdentifier]int),
	}

	resolver := &Resolver{
		client: fake.NewFakeClientWithScheme(scheme.Scheme),
		mapper: newRESTMapper(
			appsv1.SchemeGroupVersion.WithKind("Deployment"),
			corev1.SchemeGroupVersion.WithKind("Service"),
		),
		statusComputeFunc: statusComputer.Compute,
		pollInterval:      testPollInterval,
	}

	depResourceIdentifier := resourceIdentifierFromRuntimeObject(deploymentResource)
	serviceResourceIdentifier := resourceIdentifierFromRuntimeObject(serviceResource)
	identifiers := []ResourceIdentifier{
		depResourceIdentifier,
		serviceResourceIdentifier,
	}

	eventChan := resolver.WaitForStatus(context.TODO(), identifiers)

	var events []Event
	timer := time.NewTimer(testTimeout)
loop:
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				break loop
			}
			events = append(events, event)
		case <-timer.C:
			t.Fatalf("timeout waiting for resources to reach current status")
		}
	}

	expectedEvents := []struct {
		aggregateStatus status.Status
		hasResource     bool
		resourceStatus  status.Status
	}{
		{
			aggregateStatus: status.UnknownStatus,
			hasResource:     true,
			resourceStatus:  status.CurrentStatus,
		},
		{
			aggregateStatus: status.CurrentStatus,
			hasResource:     true,
			resourceStatus:  status.CurrentStatus,
		},
		{
			aggregateStatus: status.CurrentStatus,
			hasResource:     false,
		},
	}

	if want, got := len(expectedEvents), len(events); got != want {
		t.Errorf("expected %d events, but got %d", want, got)
	}

	for i, e := range events {
		ee := expectedEvents[i]
		if want, got := ee.aggregateStatus, e.AggregateStatus; got != want {
			t.Errorf("expected event %d to be %s, but got %s", i, want, got)
		}

		if ee.hasResource {
			if want, got := ee.resourceStatus, e.EventResource.Status; want != got {
				t.Errorf("expected resource event %d to be %s, but got %s", i, want, got)
			}
		}
	}
}

type statusComputer struct {
	t *testing.T

	results           map[ResourceIdentifier][]*status.Result
	resourceCallCount map[ResourceIdentifier]int
}

func (s *statusComputer) Compute(u *unstructured.Unstructured) (*status.Result, error) {
	identifier := resourceIdentifierFromRuntimeObject(u)

	resourceResults, ok := s.results[identifier]
	if !ok {
		s.t.Fatalf("No results available for resource %s/%s", u.GetNamespace(), u.GetName())
	}
	callCount := s.resourceCallCount[identifier]

	var res *status.Result
	if len(resourceResults) <= callCount {
		res = resourceResults[len(resourceResults)-1]
	} else {
		res = resourceResults[callCount]
	}
	s.resourceCallCount[identifier] = callCount + 1
	return res, nil
}

var deploymentResource = &appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "myDeployment",
		Namespace: "default",
	},
}

var statefulSetResource = &appsv1.StatefulSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "myStatefulSet",
		Namespace: "default",
	},
}

var serviceResource = &corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "myService",
		Namespace: "default",
	},
}

func newRESTMapper(gvks ...schema.GroupVersionKind) meta.RESTMapper {
	var groupVersions []schema.GroupVersion
	for _, gvk := range gvks {
		groupVersions = append(groupVersions, gvk.GroupVersion())
	}
	mapper := meta.NewDefaultRESTMapper(groupVersions)
	for _, gvk := range gvks {
		mapper.Add(gvk, meta.RESTScopeNamespace)
	}
	return mapper
}
