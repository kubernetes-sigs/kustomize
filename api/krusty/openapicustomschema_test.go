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

// Error if user tries to specify both builtin version
// and custom schema
func TestCustomOpenApiFieldBothPathAndVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- mycrd.yaml
openapi:
  version: v1.18.8
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
  version: v1.19.1
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
	assert.Equal(t, "v1191", openapi.GetSchemaVersion())
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
