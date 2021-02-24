package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
)

func TestOpenApiFieldBasicUsage(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
openapi:
  version: v1.18.8
resources:
- deployment.yaml
`)
	th.WriteF("/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	assert.Equal(t, "v1188", openapi.GetSchemaVersion())
}

func TestOpenApiFieldNotBuiltin(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
openapi:
  version: v1.14.1
resources:
- deployment.yaml
`)
	th.WriteF("/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)

	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
}

func TestOpenApiFieldDefaultVersion(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- deployment.yaml
`)
	th.WriteF("/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	assert.Equal(t, kubernetesapi.DefaultOpenAPI, openapi.GetSchemaVersion())
}

func TestOpenApiFieldFromBase(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
openapi:
  version: v1.19.0
namePrefix: a-
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("overlay", `
namePrefix: b-
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 999
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-a-myDeployment
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
`)
	assert.Equal(t, "v1190", openapi.GetSchemaVersion())
}

func TestOpenApiFieldFromOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: a-
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("overlay", `
openapi:
  version: v1.18.8
namePrefix: b-
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 999
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-a-myDeployment
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
`)
	assert.Equal(t, "v1188", openapi.GetSchemaVersion())
}

func TestOpenApiFieldOverlayTakesPrecedence(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
openapi:
  version: v1.19.0
namePrefix: a-
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("overlay", `
openapi:
  version: v1.18.8
namePrefix: b-
resources:
- ../base
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("overlay/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 999
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-a-myDeployment
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
`)
	assert.Equal(t, "v1188", openapi.GetSchemaVersion())
}

func TestOpenAPIFieldFromComponentDefault(t *testing.T) {
	input := []FileGen{writeTestBase, writeTestComponent, writeOverlayProd}
	runPath := "/prod"

	th := kusttest_test.MakeHarness(t)
	for _, f := range input {
		f(th)
	}
	th.Run(runPath, th.MakeDefaultOptions())
	assert.Equal(t, kubernetesapi.DefaultOpenAPI, openapi.GetSchemaVersion())
}

func writeTestComponentWithOlderOpenAPIVersion(th kusttest_test.Harness) {
	th.WriteC("/comp", `
openapi:
  version: v1.18.8
`)
	th.WriteF("/comp/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

const runPath = "prod"

func TestOpenAPIFieldFromComponent(t *testing.T) {
	input := []FileGen{
		writeTestBase,
		writeTestComponentWithOlderOpenAPIVersion,
		writeOverlayProd}

	th := kusttest_test.MakeHarness(t)
	for _, f := range input {
		f(th)
	}
	th.Run(runPath, th.MakeDefaultOptions())
	assert.Equal(t, "v1188", openapi.GetSchemaVersion())
}
