package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

//nolint
func writeTestSchema(th kusttest_test.Harness, filepath string) {
	th.WriteF(filepath+"mycrd_schema.json", `
{
  "definitions": {
    "v1alpha1.MyCRD": {
      "description": "MyCRD is the Schema for the mycrd API",
      "properties": {
        "apiVersion": {
          "description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
          "type": "string"
        },
        "kind": {
          "description": "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
          "type": "string"
        },
        "metadata": {
          "type": "object"
        },
        "spec": {
          "description": "MyCRDSpec defines the desired state of MyCRD",
          "properties": {
            "template": {
              "$ref": "#/definitions/io.k8s.api.core.v1.PodTemplateSpec",
              "description": "Template describes the pods that will be created."
            }
          },
          "required": [
            "template"
          ],
          "type": "object"
        },
        "status": {
          "description": "MyCRDStatus defines the observed state of MyCRD",
          "properties": {
            "success": {
              "type": "boolean"
            }
          },
          "type": "object"
        }
      },
      "type": "object",
      "x-kubernetes-group-version-kind": [
        {
          "group": "example.com",
          "kind": "MyCRD",
          "version": "v1alpha1"
        }
      ]
    }
  }
}
`)
}

// Test for issue #2825
func TestCustomOpenApiFieldNoOverlays(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
resources:
- mycrd.yaml

openapi:
  path: mycrd_schema.json

patchesStrategicMerge:
- |-
  apiVersion: example.com/v1alpha1
  kind: MyCRD
  metadata:
    name: service
  spec:
    template:
      spec:
        containers:
        - name: server
          image: nginx
`)
	th.WriteF("/app/mycrd.yaml", `
apiVersion: example.com/v1alpha1
kind: MyCRD
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - name: server
        image: server
        command: example
        ports:
        - name: grpc
          protocol: TCP
          containerPort: 8080
`)
	writeTestSchema(th, "/app/")

	// currently broken
	// desired behavior: replace image with nginx without
	// completely overwriting the containers field
	// needs to use the PodSpec merge strategy from the
	// provided openapi schema
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: example.com/v1alpha1
kind: MyCRD
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: server
`)
}
