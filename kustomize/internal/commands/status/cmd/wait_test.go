package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func TestWaitNoResources(t *testing.T) {
	inBuffer := &bytes.Buffer{}
	outBuffer := &bytes.Buffer{}

	fakeClient := &FakeClient{}

	r := GetWaitRunner()
	r.newResolverFunc = fakeResolver(fakeClient)
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err := r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}
	cleanOutput := stripansi.Strip(outBuffer.String())
	tableOutput := parseTableOutput(t, cleanOutput)

	if want, got := 2, len(tableOutput.Frames); want != got {
		t.Errorf("expected %d frames, but found %d", want, got)
	}

	aggStatuses := tableOutput.allAggStatuses()
	expectedAggStatuses := []status.Status{
		status.UnknownStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(aggStatuses, expectedAggStatuses) {
		t.Errorf("expected agg statuses to be %s, but got %s", joinStatuses(expectedAggStatuses),
			joinStatuses(aggStatuses))
	}

	resources := tableOutput.resources()
	if want, got := 0, len(resources); want != got {
		t.Errorf("expected %d resources, but found %d", want, got)
	}
}

func TestWaitMultipleUpdates(t *testing.T) {
	inBuffer := &bytes.Buffer{}
	outBuffer := &bytes.Buffer{}

	_, err := fmt.Fprint(inBuffer, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar
  namespace: default
`)
	if !assert.NoError(t, err) {
		return
	}

	fakeClient := &FakeClient{
		resourceCallbackMap: map[string]ResourceGetCallback{
			"Deployment": createDeploymentStatusFunc(),
		},
	}

	r := GetWaitRunner()
	r.newResolverFunc = fakeResolver(fakeClient, appsv1.SchemeGroupVersion.WithKind("Deployment"))
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	cleanOutput := stripansi.Strip(outBuffer.String())
	tableOutput := parseTableOutput(t, cleanOutput)

	aggStatuses := tableOutput.dedupedAggStatuses()
	expectedStatuses := []status.Status{
		status.UnknownStatus,
		status.InProgressStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(aggStatuses, expectedStatuses) {
		t.Errorf("expected deduped agg statuses to be %s, but got %s", joinStatuses(expectedStatuses),
			joinStatuses(aggStatuses))
	}

	resources := tableOutput.resources()
	if want, got := 1, len(resources); got != want {
		t.Errorf("expected %d resource, but got %d", want, got)
	}

	resource := resources[0]
	resourceStatuses := tableOutput.dedupedStatusesForResource(resource)
	expectedResourceStatuses := []status.Status{
		status.InProgressStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(expectedResourceStatuses, resourceStatuses) {
		t.Errorf("expected resource %q to have statuses %s, but got %s", resource.name,
			joinStatuses(expectedResourceStatuses), joinStatuses(resourceStatuses))
	}
}

func TestWaitMultipleResources(t *testing.T) {
	inBuffer := &bytes.Buffer{}
	outBuffer := &bytes.Buffer{}

	_, err := fmt.Fprint(inBuffer, `
apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: foo
    namespace: default
- apiVersion: v1
  kind: Service
  metadata:
    name: bar
    namespace: default
`)
	if !assert.NoError(t, err) {
		return
	}

	fakeClient := &FakeClient{
		resourceCallbackMap: map[string]ResourceGetCallback{
			"Pod":     createPodStatusFunc(),
			"Service": createServiceStatusFunc(),
		},
	}

	r := GetWaitRunner()
	r.newResolverFunc = fakeResolver(fakeClient, corev1.SchemeGroupVersion.WithKind("Pod"),
		corev1.SchemeGroupVersion.WithKind("Service"))
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	cleanOutput := stripansi.Strip(outBuffer.String())
	tableOutput := parseTableOutput(t, cleanOutput)

	aggStatuses := tableOutput.dedupedAggStatuses()
	if want, got := status.CurrentStatus, aggStatuses[len(aggStatuses)-1]; want != got {
		t.Errorf("expected final agg statuses to be %s, but got %s", want, got)
	}

	resources := tableOutput.resources()
	if want, got := 2, len(resources); got != want {
		t.Errorf("expected %d resource, but got %d", want, got)
	}

	for _, resource := range resources {
		resourceStatuses := tableOutput.dedupedStatusesForResource(resource)
		if want, got := status.CurrentStatus, resourceStatuses[len(resourceStatuses)-1]; want != got {
			t.Errorf("expected resource %q to have final status %s, but got %s", resource.name, want, got)
		}
	}
}
