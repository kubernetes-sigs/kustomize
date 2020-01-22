package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/kustomize/kstatus/status"
)

func TestEmptyManifest(t *testing.T) {
	inBuffer := &bytes.Buffer{}
	outBuffer := &bytes.Buffer{}

	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := GetFetchRunner()
	r.newResolverFunc = fakeResolver(fakeClient)
	r.Command.SetArgs([]string{})
	r.Command.SetIn(inBuffer)
	r.Command.SetOut(outBuffer)

	err := r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	output := outBuffer.String()
	lines := strings.Split(output, "\n")

	if want, got := 2, len(lines); want != got {
		t.Errorf("Expected %d lines, but got %d", want, got)
	}
}

func TestFetchStatusFromManifestStdIn(t *testing.T) {
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

	deployment := createDeployment("bar", "default", 42, appsv1.DeploymentStatus{
		ObservedGeneration: 1,
	})

	fakeClient := fake.NewFakeClientWithScheme(scheme, deployment)

	r := GetFetchRunner()
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

	expectedResource := ResourceIdentifier{
		apiVersion: "apps",
		kind:       "Deployment",
		namespace:  "default",
		name:       "bar",
	}
	expectedStatus := status.InProgressStatus
	expectedMessage := "Deployment generation is 2, but latest observed generation is 1"

	verifyOutputContains(t, tableOutput, expectedResource, expectedStatus, expectedMessage)
}

//nolint:funlen
func TestFetchStatusFromManifestsFiles(t *testing.T) {
	d, err := ioutil.TempDir("", "status-fetch-test")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(d)

	err = ioutil.WriteFile(filepath.Join(d, "dep.yaml"), []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: default
`), 0600)
	if !assert.NoError(t, err) {
		return
	}
	err = ioutil.WriteFile(filepath.Join(d, "svc.yaml"), []byte(`
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: default
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	replicas := int32(42)
	deployment := createDeployment("foo", "default", replicas, appsv1.DeploymentStatus{
		ObservedGeneration: 2,
		Replicas:           replicas,
		ReadyReplicas:      replicas,
		AvailableReplicas:  replicas,
		UpdatedReplicas:    replicas,
		Conditions: []appsv1.DeploymentCondition{
			{
				Type:   appsv1.DeploymentAvailable,
				Status: v1.ConditionTrue,
			},
		},
	})
	service := createService("foo", "default")

	fakeClient := fake.NewFakeClientWithScheme(scheme, deployment, service)

	outBuffer := &bytes.Buffer{}

	r := GetFetchRunner()
	r.newResolverFunc = fakeResolver(fakeClient, appsv1.SchemeGroupVersion.WithKind("Deployment"),
		v1.SchemeGroupVersion.WithKind("Service"))
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(outBuffer)

	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	cleanOutput := stripansi.Strip(outBuffer.String())
	tableOutput := parseTableOutput(t, cleanOutput)

	expectedDeploymentResource := ResourceIdentifier{
		apiVersion: "apps",
		kind:       "Deployment",
		namespace:  "default",
		name:       "foo",
	}
	expectedDeploymentStatus := status.CurrentStatus
	expectedDeploymentMessage := "Deployment is available. Replicas: 42"
	verifyOutputContains(t, tableOutput, expectedDeploymentResource, expectedDeploymentStatus, expectedDeploymentMessage)

	expectedServiceResource := ResourceIdentifier{
		apiVersion: "",
		kind:       "Service",
		namespace:  "default",
		name:       "foo",
	}
	expectedServiceStatus := status.CurrentStatus
	expectedServiceMessage := "Service is ready"

	verifyOutputContains(t, tableOutput, expectedServiceResource, expectedServiceStatus, expectedServiceMessage)
}

func createDeployment(name, namespace string, replicas int32, status appsv1.DeploymentStatus) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 2,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: status,
	}
}

func createService(name, namespace string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func verifyOutputContains(t *testing.T, tableOutput TableOutput, resource ResourceIdentifier, status status.Status, message string) {
	if len(tableOutput.Frames) == 0 {
		t.Fatalf("expected match for resource %s, but output had no frames", resource.name)
	}
	firstFrame := tableOutput.Frames[0]
	var foundResource ResourceOutput
	match := false
	for _, resourceOutput := range firstFrame.Resources {
		if resourceOutput.identifier.Equals(resource) {
			foundResource = resourceOutput
			match = true
			break
		}
	}
	if !match {
		t.Errorf("expected match for resource %s, but didn't find it", resource.name)
	}
	if want, got := status, foundResource.status; want != got {
		t.Errorf("expected status %s for resource %s, but got %s", want, resource.name, got)
	}
	if want, got := message, foundResource.message; !strings.HasPrefix(want, got) {
		t.Errorf("expected message %s for resource %s, but got %s", want, resource.name, got)
	}
}
