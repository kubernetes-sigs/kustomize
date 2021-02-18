package krusty_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/openapi"
)

func writeTestSchema(th kusttest_test.Harness, filepath string) {
	bytes, _ := ioutil.ReadFile("testdata/customschema.json")
	th.WriteF(filepath+"mycrd_schema.json", string(bytes))
}

func writeTestComponentWithCustomSchema(th kusttest_test.Harness) {
	writeTestSchema(th, "/app/comp/")
	openapi.ResetOpenAPI()
	th.WriteC("/app/comp", `
openapi:
  path: mycrd_schema.json
`)
	th.WriteF("/app/comp/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

// Test for issue #2825
func TestCustomOpenApiFieldBasicUsage(t *testing.T) {
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
	openapi.ResetOpenAPI()

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
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
`)
}

// Error if user tries to specify both builtin version
// and custom schema
func TestCustomOpenApiFieldBothPathAndVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
resources:
- mycrd.yaml

openapi:
  version: v1.18.8
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
	openapi.ResetOpenAPI()
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	assert.Error(t, err)
	assert.Equal(t,
		"builtin version and custom schema provided, cannot use both",
		err.Error())
}

// Test for if the filepath specified is not found
func TestCustomOpenApiFieldFileNotFound(t *testing.T) {
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
	openapi.ResetOpenAPI()
	err := th.RunWithErr("/app", th.MakeDefaultOptions())
	assert.Error(t, err)
	assert.Equal(t,
		"'/app/mycrd_schema.json' doesn't exist",
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
	th.WriteF("base/mycrd.yaml", `
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
	th.WriteK("overlay", `
resources:
- ../base

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
	writeTestSchema(th, "/base/")
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
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
`)
	assert.Equal(t, "using custom schema from file provided",
		openapi.GetSchemaVersion())
}

func TestCustomOpenApiFieldFromOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- mycrd.yaml
`)
	th.WriteF("base/mycrd.yaml", `
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
	th.WriteK("overlay", `
resources:
- ../base
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
	writeTestSchema(th, "/overlay/")
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
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
`)
	assert.Equal(t, "using custom schema from file provided",
		openapi.GetSchemaVersion())
}

func TestCustomOpenApiFieldOverlayTakesPrecedence(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- mycrd.yaml
openapi:
  path: mycrd_schema.json
`)
	th.WriteF("base/mycrd.yaml", `
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
	th.WriteK("overlay", `
resources:
- ../base
openapi:
  version: v1.19.1
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
	writeTestSchema(th, "/base/")
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
	assert.Equal(t, "v1191",
		openapi.GetSchemaVersion())
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
	th.Run(runPath, th.MakeDefaultOptions())
	assert.Equal(t, "using custom schema from file provided", openapi.GetSchemaVersion())
}
