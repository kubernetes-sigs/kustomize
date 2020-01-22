package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

func TestEventsNoResources(t *testing.T) {
	inBuffer := &bytes.Buffer{}
	outBuffer := &bytes.Buffer{}

	fakeClient := &FakeClient{}

	r := GetEventsRunner()
	r.newResolverFunc = fakeResolver(fakeClient)
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err := r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	eventOutput := parseEventOutput(t, outBuffer.String())

	if want, got := 1, len(eventOutput.events); want != got {
		t.Errorf("expected %d events, but got %d", want, got)
	}

	event := eventOutput.events[0]
	if want, got := status.CurrentStatus, event.aggStatus; want != got {
		t.Errorf("expected agg status %s, but got %s", want, got)
	}
}

func TestEventsMultipleUpdates(t *testing.T) {
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

	r := GetEventsRunner()
	r.newResolverFunc = fakeResolver(fakeClient, appsv1.SchemeGroupVersion.WithKind("Deployment"))
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	eventOutput := parseEventOutput(t, outBuffer.String())

	aggStatuses := eventOutput.allAggStatuses()
	expectedAggStatuses := []status.Status{
		status.InProgressStatus,
		status.InProgressStatus,
		status.InProgressStatus,
		status.InProgressStatus,
		status.CurrentStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(aggStatuses, expectedAggStatuses) {
		t.Errorf("expected agg statuses to be %s, but got %s", joinStatuses(expectedAggStatuses),
			joinStatuses(aggStatuses))
	}

	resources := eventOutput.allResources()
	if want, got := 1, len(resources); want != got {
		t.Errorf("expected %d resource, but got %d", want, got)
	}

	resource := resources[0]
	resourceStatuses := eventOutput.statusesForResource(resource)
	expectedResourceStatuses := []status.Status{
		status.InProgressStatus,
		status.InProgressStatus,
		status.InProgressStatus,
		status.InProgressStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(resourceStatuses, expectedResourceStatuses) {
		t.Errorf("expected statuses to be %s, but got %s", joinStatuses(expectedResourceStatuses),
			joinStatuses(resourceStatuses))
	}
}

func TestEventsMultipleResources(t *testing.T) {
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

	r := GetEventsRunner()
	r.newResolverFunc = fakeResolver(fakeClient, corev1.SchemeGroupVersion.WithKind("Pod"),
		corev1.SchemeGroupVersion.WithKind("Service"))
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	eventOutput := parseEventOutput(t, outBuffer.String())

	aggStatuses := eventOutput.allAggStatuses()
	expectedAggStatuses := []status.Status{
		status.UnknownStatus,
		status.CurrentStatus,
		status.CurrentStatus,
	}
	if !reflect.DeepEqual(aggStatuses, expectedAggStatuses) {
		t.Errorf("expected agg statuses to be %s, but got %s", joinStatuses(expectedAggStatuses),
			joinStatuses(aggStatuses))
	}

	resources := eventOutput.allResources()
	if want, got := 2, len(resources); got != want {
		t.Errorf("expected %d resource, but got %d", want, got)
	}

	for _, resource := range resources {
		resourceStatuses := eventOutput.statusesForResource(resource)
		if want, got := status.CurrentStatus, resourceStatuses[len(resourceStatuses)-1]; want != got {
			t.Errorf("expected resource %q to have final status %s, but got %s", resource.name, want, got)
		}
	}
}

type EventOutput struct {
	events       []EventOutputLine
	unknownLines []string
}

func (e *EventOutput) allAggStatuses() []status.Status {
	var aggStatuses []status.Status
	for _, event := range e.events {
		aggStatuses = append(aggStatuses, event.aggStatus)
	}
	return aggStatuses
}

func (e *EventOutput) allResources() []ResourceIdentifier {
	var resources []ResourceIdentifier
	seenResources := make(map[ResourceIdentifier]bool)
	for _, event := range e.events {
		if !event.isResourceUpdateEvent() {
			continue
		}
		r := event.identifier
		if _, found := seenResources[r]; !found {
			resources = append(resources, r)
			seenResources[r] = true
		}
	}
	return resources
}

func (e *EventOutput) statusesForResource(resource ResourceIdentifier) []status.Status {
	var statuses []status.Status
	for _, event := range e.events {
		if !event.isResourceUpdateEvent() {
			continue
		}
		if event.identifier.Equals(resource) {
			statuses = append(statuses, event.status)
		}
	}
	return statuses
}

type EventOutputLine struct {
	eventType  string
	aggStatus  status.Status
	identifier ResourceIdentifier
	status     status.Status
	message    string
}

func (e *EventOutputLine) isResourceUpdateEvent() bool {
	return e.eventType == string(wait.ResourceUpdate)
}

var (
	eventRegex = regexp.MustCompile(`^\s*` +
		`(?P<eventType>\S+)\s+` +
		`(?P<aggStatus>\S+)\s+` +
		`((?P<resourceType>\S+)\s+` +
		`(?P<namespace>\S+)\s+` +
		`(?P<name>\S+)\s+` +
		`(?P<status>\S+)\s+` +
		`(?P<message>.*\S)){0,1}` +
		`\s*$`)
	eventHeaderRegex = regexp.MustCompile(`^\s*` +
		`EVENT TYPE\s+` +
		`AGG STATUS\s+` +
		`TYPE\s+` +
		`NAMESPACE\s+` +
		`NAME\s+` +
		`STATUS\s+` +
		`MESSAGE` +
		`\s*$`)
)

func parseEventOutput(_ *testing.T, output string) EventOutput {
	var eventOutput EventOutput
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue // Ignore empty lines
		}
		match := eventHeaderRegex.FindStringSubmatch(line)
		if match != nil {
			continue // Ignore headers
		}
		match = eventRegex.FindStringSubmatch(line)
		if match == nil {
			eventOutput.unknownLines = append(eventOutput.unknownLines, line)
			continue
		}

		eventOutputLine := EventOutputLine{
			eventType: match[1],
			aggStatus: status.FromStringOrDie(match[2]),
		}

		if eventOutputLine.eventType == string(wait.ResourceUpdate) {
			resourceType := match[4]
			parts := strings.Split(resourceType, "/")
			var identifier ResourceIdentifier
			if len(parts) == 2 {
				identifier.apiVersion = parts[0]
				identifier.kind = parts[1]
			} else {
				identifier.apiVersion = strings.Join(parts[:2], "/")
				identifier.kind = parts[2]
			}
			identifier.namespace = match[5]
			identifier.name = match[6]
			eventOutputLine.identifier = identifier
			eventOutputLine.status = status.FromStringOrDie(match[7])
			eventOutputLine.message = match[8]
		}

		eventOutput.events = append(eventOutput.events, eventOutputLine)
	}
	return eventOutput
}
