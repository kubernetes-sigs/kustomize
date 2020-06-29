package fnplugin

import (
	"bytes"
	"fmt"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"testing"
)

func TestToResourceList(t *testing.T) {
	in := []byte(`apiVersion: v1
data:
  key1: oldValue
kind: ConfigMap
metadata:
  annotations:
  kustomize.config.k8s.io/id: |
    kind: ConfigMap
    name: config1
    version: v1
  name: config1
---
apiVersion: v1
data:
  key1: oldValue
kind: ConfigMap
metadata:
  annotations:
  kustomize.config.k8s.io/id: |
    kind: ConfigMap
    name: config2
    version: v1
  name: config2`)
	outBuffer, err := toResourceList(in)
	if err != nil {
		t.Fatal(err)
	}

	expected := `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: v1
  data:
    key1: oldValue
  kind: ConfigMap
  metadata:
    kustomize.config.k8s.io/id: |
      kind: ConfigMap
      name: config1
      version: v1
    name: config1
- apiVersion: v1
  data:
    key1: oldValue
  kind: ConfigMap
  metadata:
    kustomize.config.k8s.io/id: |
      kind: ConfigMap
      name: config2
      version: v1
    name: config2
`

	if outBuffer.String() != expected {
		t.Fatalf("output \n%s\n doesn't match expected \n%s\n", outBuffer.String(), expected)
	}
}

func TestToResourceListWithEmptyInput(t *testing.T) {
	expected := fmt.Sprintf("apiVersion: %s\nkind: %s", kio.ResourceListAPIVersion, kio.ResourceListKind)
	outBuffer, err := toResourceList(nil)
	if err != nil {
		t.Fatal(err)
	}
	if outBuffer.String() != expected {
		t.Fatalf("output \n%s\n doesn't match expected \n%s\n", outBuffer.String(), expected)
	}
}

func TestInjectFunctionConfig(t *testing.T) {
	input := []byte(`apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList`)
	functionConfig, err := bytesToRNode([]byte(`apiVersion: foo-corp.com/v1
kind: FulfillmentCenter
metadata:
  name: staging
  metadata:
    annotations:
      config.k8s.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
spec:
  address: "100 Main St."`))
	if err != nil {
		t.Fatal(err)
	}
	inputBuffer := bytes.Buffer{}
	inputBuffer.Write(input)
	err = injectFunctionConfig(&inputBuffer, functionConfig)
	if err != nil {
		t.Fatal(err)
	}
	expected := `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    metadata:
      annotations:
        config.k8s.io/function: |
          container:
            image: gcr.io/example/foo:v1.0.0
  spec:
    address: "100 Main St."
`

	if inputBuffer.String() != expected {
		t.Fatalf("output \n%s\n doesn't match expected \n%s\n", inputBuffer.String(), expected)
	}
}
