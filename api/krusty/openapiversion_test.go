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
  version: v1.20.4
resources:
- deployment.yaml
`)
	th.WriteF("deployment.yaml", `
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
	assert.Equal(t, "v1204", openapi.GetSchemaVersion())
}

func TestOpenApiFieldNotBuiltin(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
openapi:
  version: v1.14.1
resources:
- deployment.yaml
`)
	th.WriteF("deployment.yaml", `
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
	th.WriteF("deployment.yaml", `
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
