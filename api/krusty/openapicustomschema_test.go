// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func writeTestSchema(th kusttest_test.Harness, filepath string) {
	bytes, _ := ioutil.ReadFile("testdata/customschema.json")
	th.WriteF(filepath+"mycrd_schema.json", string(bytes))
}

func writeTestSchemaYaml(th kusttest_test.Harness, filepath string) {
	bytes, _ := ioutil.ReadFile("testdata/customschema.yaml")
	th.WriteF(filepath+"mycrd_schema.yaml", string(bytes))
}

func writeCustomResource(th kusttest_test.Harness, filepath string) {
	th.WriteF(filepath, `
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
}

func writeOtherCustomResource(th kusttest_test.Harness, filepath string) {
	th.WriteF(filepath, `
apiVersion: crd.com/v1alpha1
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
}

func writeTestComponentWithCustomSchema(th kusttest_test.Harness) {
	writeTestSchema(th, "comp/")
	openapi.ResetOpenAPI()
	th.WriteC("comp", `
openapi:
  path: mycrd_schema.json
`)
	th.WriteF("comp/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

const customSchemaPatch = `
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
`

const customSchemaPatchMultipleGvks = `
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
- |-
  apiVersion: crd.com/v1alpha1
  kind: MyCRD
  metadata:
    name: service
  spec:
    template:
      spec:
        containers:
        - name: server
          image: nginx
`

const patchedCustomResource = `
apiVersion: example.com/v1alpha1
kind: MyCRD
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
`

// Test for issue #2825
func TestCustomOpenApiFieldBasicUsage(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.json
`+customSchemaPatch)
	writeCustomResource(th, "mycrd.yaml")
	writeTestSchema(th, "./")
	openapi.ResetOpenAPI()
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, patchedCustomResource)
}

func TestCustomOpenApiFieldWithTwoGvks(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
- myothercrd.yaml
openapi:
  path: mycrd_schema.json
`+customSchemaPatchMultipleGvks)
	writeCustomResource(th, "mycrd.yaml")
	writeOtherCustomResource(th, "myothercrd.yaml")
	writeTestSchema(th, "./")
	openapi.ResetOpenAPI()
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: example.com/v1alpha1
kind: MyCRD
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
---
apiVersion: crd.com/v1alpha1
kind: MyCRD
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
`)
}

func TestCustomOpenApiFieldYaml(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.yaml
`+customSchemaPatch)
	writeCustomResource(th, "mycrd.yaml")
	writeTestSchemaYaml(th, "./")
	openapi.ResetOpenAPI()
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, patchedCustomResource)
}

// Error if user tries to specify both builtin version
// and custom schema
func TestCustomOpenApiFieldBothPathAndVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
openapi:
  version: v1.21.2
  path: mycrd_schema.json
`+customSchemaPatch)
	writeCustomResource(th, "mycrd.yaml")
	writeTestSchema(th, "./")
	openapi.ResetOpenAPI()
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	assert.Error(t, err)
	assert.Equal(t,
		"builtin version and custom schema provided, cannot use both",
		err.Error())
}

// Test for if the filepath specified is not found
func TestCustomOpenApiFieldFileNotFound(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.json
`+customSchemaPatch)
	writeCustomResource(th, "mycrd.yaml")
	openapi.ResetOpenAPI()
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	assert.Error(t, err)
	assert.Equal(t,
		"'/mycrd_schema.json' doesn't exist",
		err.Error())
}

func TestCustomOpenApiFieldFromBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.json
`)
	th.WriteK("overlay", `
resources:
- ../base
`+customSchemaPatch)
	writeCustomResource(th, "base/mycrd.yaml")
	writeTestSchema(th, "base/")
	openapi.ResetOpenAPI()
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, patchedCustomResource)
	assert.Equal(t, "using custom schema from file provided",
		openapi.GetSchemaVersion())
}

func TestCustomOpenApiFieldFromOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- mycrd.yaml
`)
	th.WriteK("overlay", `
resources:
- ../base
openapi:
  path: mycrd_schema.json
`+customSchemaPatch)
	writeCustomResource(th, "base/mycrd.yaml")
	writeTestSchema(th, "overlay/")
	openapi.ResetOpenAPI()
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, patchedCustomResource)
	assert.Equal(t, "using custom schema from file provided",
		openapi.GetSchemaVersion())
}

func TestCustomOpenApiFieldOverlayTakesPrecedence(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	openapi.ResetOpenAPI()
	th.WriteK("base", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.json
`)
	th.WriteK("overlay", `
resources:
- ../base
openapi:
  version: v1.21.2
`+customSchemaPatch)
	writeCustomResource(th, "base/mycrd.yaml")
	writeTestSchema(th, "base/")
	openapi.ResetOpenAPI()
	m := th.Run("overlay", th.MakeDefaultOptions())
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
	assert.Equal(t, "v1212", openapi.GetSchemaVersion())
}

func TestCustomOpenAPIFieldFromComponent(t *testing.T) {
	input := []FileGen{
		writeTestBase,
		writeTestComponentWithCustomSchema,
		writeOverlayProd}

	th := kusttest_test.MakeHarness(t)
	for _, f := range input {
		f(th)
	}
	openapi.ResetOpenAPI()
	th.Run("prod", th.MakeDefaultOptions())
	assert.Equal(t, "using custom schema from file provided", openapi.GetSchemaVersion())
}

// test for https://github.com/kubernetes-sigs/kustomize/issues/4179
// kustomize is not seeing the openapi field from the component defined in the overlay
func TestCustomOpenAPIFieldFromComponentWithOverlays(t *testing.T) {
	if val, ok := os.LookupEnv("OPENAPI_TEST"); !ok || val != "true" {
		t.SkipNow()
	}

	th := kusttest_test.MakeHarness(t)

	// overlay declaring the component
	th.WriteK("overlays/overlay-component-openapi", `resources:
- ../base/
components:
- ../../components/dc-openapi
`)

	// base kustomization
	th.WriteK("overlays/base", `resources:
- dc.yml
`)

	// resource declared in the base kustomization
	th.WriteF("overlays/base/dc.yml", `apiVersion: apps.openshift.io/v1
kind: DeploymentConfig
metadata:
  name: my-dc
spec:
  template:
    spec:
      initContainers:
        - name: init
      containers:
        - name: container
          env:
            - name: foo
              value: bar
          volumeMounts:
            - name: cm
              mountPath: /opt/cm
      volumes:
        - name: cm
          configMap:
            name: cm
`)

	// openapi schema referred to by the component
	bytes, _ := ioutil.ReadFile("testdata/openshiftschema.json")
	th.WriteF("components/dc-openapi/openapi.json", string(bytes))

	// patch referred to by the component
	th.WriteF("components/dc-openapi/patch.yml", `apiVersion: apps.openshift.io/v1
kind: DeploymentConfig
metadata:
  name: my-dc
spec:
  template:
    spec:
      containers:
        - name: container
          volumeMounts:
            - name: additional-cm
              mountPath: /mnt
      volumes:
        - name: additional-cm
          configMap:
             name: additional-cm
`)

	// component declared in overlay with custom schema and patch
	th.WriteC("components/dc-openapi", `patches:
  - patch.yml
openapi:
  path: openapi.json
`)

	openapi.ResetOpenAPI()
	m := th.Run("overlays/overlay-component-openapi", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: apps.openshift.io/v1
kind: DeploymentConfig
metadata:
  name: my-dc
spec:
  template:
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        name: container
        volumeMounts:
        - mountPath: /mnt
          name: additional-cm
        - mountPath: /opt/cm
          name: cm
      initContainers:
      - name: init
      volumes:
      - configMap:
          name: additional-cm
        name: additional-cm
      - configMap:
          name: cm
        name: cm
`)
}
