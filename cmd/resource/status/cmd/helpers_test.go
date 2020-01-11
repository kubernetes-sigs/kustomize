package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

type TableOutput struct {
	Frames      []TableOutputFrame
	UnknownRows []string
}

func (t *TableOutput) allAggStatuses() []status.Status {
	var statuses []status.Status
	for _, frame := range t.Frames {
		if frame.AggregateStatus != "" {
			statuses = append(statuses, frame.AggregateStatus)
		}
	}
	return statuses
}

func (t *TableOutput) dedupedAggStatuses() []status.Status {
	var dedupedStatuses []status.Status
	statuses := t.allAggStatuses()
	var previousStatus status.Status
	for _, s := range statuses {
		if s != previousStatus {
			dedupedStatuses = append(dedupedStatuses, s)
			previousStatus = s
		}
	}
	return dedupedStatuses
}

func (t *TableOutput) resources() []ResourceIdentifier {
	seenResources := make(map[ResourceIdentifier]bool)
	var resources []ResourceIdentifier
	for _, frame := range t.Frames {
		for _, resource := range frame.Resources {
			r := resource.identifier
			_, found := seenResources[r]
			if !found {
				seenResources[r] = true
				resources = append(resources, r)
			}
		}
	}
	return resources
}

func (t *TableOutput) dedupedStatusesForResource(resource ResourceIdentifier) []status.Status {
	var dedupedStatuses []status.Status
	var previousStatus status.Status
	for _, frame := range t.Frames {
		for _, r := range frame.Resources {
			if r.identifier.Equals(resource) {
				if r.status != previousStatus {
					previousStatus = r.status
					dedupedStatuses = append(dedupedStatuses, r.status)
				}
			}
		}
	}
	return dedupedStatuses
}

type TableOutputFrame struct {
	AggregateStatus status.Status
	Resources       []ResourceOutput
}

type ResourceIdentifier struct {
	apiVersion string
	kind       string
	name       string
	namespace  string
}

func (r ResourceIdentifier) Equals(identifier ResourceIdentifier) bool {
	return r.apiVersion == identifier.apiVersion &&
		r.kind == identifier.kind &&
		r.namespace == identifier.namespace &&
		r.name == identifier.name
}

type ResourceOutput struct {
	identifier ResourceIdentifier
	status     status.Status
	message    string
}

var (
	headerRegex    = regexp.MustCompile(`^\s*TYPE\s+NAMESPACE\s+NAME\s+STATUS\s+MESSAGE\s*$`)
	resourceRegex  = regexp.MustCompile(`^(?P<resourceType>\S+)\s+(?P<namespace>\S+)\s+(?P<name>\S+)\s+(?P<status>\S+)\s+(?P<message>.*\S)\s*$`)
	aggStatusRegex = regexp.MustCompile(`^\s*AggregateStatus: (?P<aggregateStatus>\S+)\s*$`)
)

func parseTableOutput(_ *testing.T, output string) TableOutput {
	tableOutput := TableOutput{}

	lines := strings.Split(output, "\n")

	hasAggStatus := false
	var currentFrame TableOutputFrame
	for i, line := range lines {
		if len(line) == 0 {
			continue // We don't care about empty lines.
		}

		// Check for lines with aggregate status. They are not always present, but if they are,
		// they always start a new frame of output.
		match := aggStatusRegex.FindStringSubmatch(line)
		if match != nil {
			hasAggStatus = true
			if i != 0 {
				tableOutput.Frames = append(tableOutput.Frames, currentFrame)
			}
			currentFrame = TableOutputFrame{
				AggregateStatus: status.FromStringOrDie(match[1]),
			}
			continue
		}

		match = headerRegex.FindStringSubmatch(line)
		if match != nil {
			if !hasAggStatus {
				if i != 0 {
					tableOutput.Frames = append(tableOutput.Frames, currentFrame)
				}
				currentFrame = TableOutputFrame{}
			}
			continue
		}

		match = resourceRegex.FindStringSubmatch(line)
		if match != nil {
			var identifier ResourceIdentifier
			resourceType := match[1]
			parts := strings.Split(resourceType, "/")
			if len(parts) == 2 {
				identifier.apiVersion = parts[0]
				identifier.kind = parts[1]
			} else {
				identifier.apiVersion = strings.Join(parts[:2], "/")
				identifier.kind = parts[2]
			}
			identifier.namespace = match[2]
			identifier.name = match[3]

			res := ResourceOutput{
				identifier: identifier,
			}
			res.status = status.FromStringOrDie(match[4])
			res.message = match[5]
			currentFrame.Resources = append(currentFrame.Resources, res)
			continue
		}
		tableOutput.UnknownRows = append(tableOutput.UnknownRows, line)
	}
	tableOutput.Frames = append(tableOutput.Frames, currentFrame)
	return tableOutput
}

func createDeploymentStatusFunc() func(*unstructured.Unstructured) error {
	metadataMap := map[string]interface{}{
		"generation": int64(2),
	}
	specMap := map[string]interface{}{
		"replicas": int64(2),
	}
	statusMap := map[string]interface{}{
		"observedGeneration": int64(2),
		"replicas":           int64(4),
		"updatedReplicas":    int64(4),
		"readyReplicas":      int64(4),
	}
	var conditions = make([]interface{}, 0)
	conditions = append(conditions, map[string]interface{}{
		"type":   "Available",
		"status": "True",
	})
	callbackCount := int64(0)
	return func(deployment *unstructured.Unstructured) error {
		_ = unstructured.SetNestedMap(deployment.Object, metadataMap, "metadata")
		_ = unstructured.SetNestedMap(deployment.Object, specMap, "spec")
		statusMap["availableReplicas"] = callbackCount
		_ = unstructured.SetNestedMap(deployment.Object, statusMap, "status")
		_ = unstructured.SetNestedSlice(deployment.Object, conditions, "status", "conditions")
		callbackCount++
		return nil
	}
}

func createServiceStatusFunc() func(*unstructured.Unstructured) error {
	return func(*unstructured.Unstructured) error {
		return nil
	}
}

func createPodStatusFunc() func(*unstructured.Unstructured) error {
	statusMap := map[string]interface{}{
		"phase": "Succeeded",
	}
	return func(pod *unstructured.Unstructured) error {
		_ = unstructured.SetNestedMap(pod.Object, statusMap, "status")
		return nil
	}
}

type ResourceGetCallback func(resource *unstructured.Unstructured) error

type FakeClient struct {
	resourceCallbackMap map[string]ResourceGetCallback
}

func (f *FakeClient) Get(_ context.Context, _ client.ObjectKey, obj runtime.Object) error {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	callbackFunc, found := f.resourceCallbackMap[kind]
	if !found {
		return fmt.Errorf("no callback func found for kind %s", kind)
	}
	u := obj.(*unstructured.Unstructured)
	return callbackFunc(u)
}

func (f *FakeClient) List(context.Context, runtime.Object, ...client.ListOption) error {
	return nil
}

func fakeResolver(fakeClient client.Reader, mapperTypes ...schema.GroupVersionKind) newResolverFunc {
	return func(pollInterval time.Duration) (*wait.Resolver, error) {
		var groupVersions []schema.GroupVersion
		for _, gvk := range mapperTypes {
			groupVersions = append(groupVersions, gvk.GroupVersion())
		}
		mapper := meta.NewDefaultRESTMapper(groupVersions)
		for _, gvk := range mapperTypes {
			mapper.Add(gvk, meta.RESTScopeNamespace)
		}

		return wait.NewResolver(fakeClient, mapper, pollInterval), nil
	}
}

func joinStatuses(statuses []status.Status) string {
	var stringStatuses []string
	for _, s := range statuses {
		stringStatuses = append(stringStatuses, s.String())
	}
	return strings.Join(stringStatuses, ",")
}
