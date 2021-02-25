package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Checks that a patch at the top of the stack can refer to resources
// by an intermediate name after it has gone through multiple name
// transformations.
// Ref: Issue #3455
func TestIntermediateName(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("gcp", `
namePrefix: gcp-
resources:
- ../emea
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("gcp/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prod-foo
spec:
  replicas: 999
`)
	th.WriteK("emea", `
namePrefix: emea-
resources:
- ../prod
`)
	th.WriteK("prod", `
namePrefix: prod-
resources:
- ../base
`)
	th.WriteK("base", `
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	m := th.Run("gcp", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gcp-emea-prod-foo
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
`)
}

// Tests that if resources in different layers (containing name
// transformations) have the same name, there is no conflict
func TestIntermediateNameSameNameDifferentLayer(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("gcp", `
namePrefix: gcp-
resources:
- ../emea
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("gcp/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prod-foo
spec:
  replicas: 999
`)
	th.WriteK("emea", `
namePrefix: emea-
resources:
- ../prod
- deployment.yaml
`)
	th.WriteF("emea/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("prod", `
namePrefix: prod-
resources:
- ../base
`)
	th.WriteK("base", `
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	m := th.Run("gcp", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gcp-emea-prod-foo
spec:
  replicas: 999
  template:
    spec:
      containers:
      - image: whatever
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gcp-emea-foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
}

// Same as above test but tries to refer to the name foo
// instead of prod-foo
func TestIntermediateNameAmbiguous(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("gcp", `
namePrefix: gcp-
resources:
- ../emea
patchesStrategicMerge:
- depPatch.yaml
`)
	th.WriteF("gcp/depPatch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  replicas: 999
`)
	th.WriteK("emea", `
namePrefix: emea-
resources:
- ../prod
- deployment.yaml
`)
	th.WriteF("emea/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	th.WriteK("prod", `
namePrefix: prod-
resources:
- ../base
`)
	th.WriteK("base", `
resources:
- deployment.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
      - image: whatever
`)
	err := th.RunWithErr("gcp", th.MakeDefaultOptions())
	assert.Error(t, err)
}

// Test for issue #3228
// References to resources by their intermediate names after multiple
// name transformations should be supported
func TestIntermediateNameSecretKeyRefDiamond(t *testing.T) {
	th := kusttest_test.MakeHarness(t)

	th.WriteK(".", `
namePrefix: project-
resources:
- app`)

	th.WriteK("app", `
resources:
- resources/deployment.yaml
- resources/xql
`)

	th.WriteK("app/resources/xql", `
resources:
- xql-zero
- xql-one
configurations:
- ./kustomizeconfig.yaml
`)

	th.WriteF("app/resources/xql/kustomizeconfig.yaml", `
varReference:
- path: spec/template/spec/containers/env/valueFrom/secretKeyRef/name
`)

	th.WriteK("app/resources/xql/xql-one", `
namePrefix: xql-one-
resources:
- ../../../../bases/xql
secretGenerator:
- name: xql-secret
  behavior: merge
  envs:
  - config/xql-one-secret.env
vars:
- name: PROJECT_XQL_ONE_SECRET_NAME
  objref:
    kind: Secret
    name: xql-secret
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
`)

	th.WriteF("app/resources/xql/xql-one/config/xql-one-secret.env", `
arg=1
`)

	th.WriteK("app/resources/xql/xql-zero", `
namePrefix: xql-zero-
resources:
- ../../../../bases/xql
secretGenerator:
- name: xql-secret
  behavior: merge
  envs:
  - config/xql-zero-secret.env
vars:
- name: PROJECT_XQL_ZERO_SECRET_NAME
  objref:
    kind: Secret
    name: xql-secret
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
`)

	th.WriteF("app/resources/xql/xql-zero/config/xql-zero-secret.env", `
arg=0
`)

	th.WriteF("app/resources/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: app
        image: example.com/app:latest
        imagePullPolicy: Always
        env:
        - name: XQL_ZERO_ARG
          valueFrom:
            secretKeyRef:
              name: $(PROJECT_XQL_ZERO_SECRET_NAME)
              key: arg
        - name: XQL_ZERO_PASSWORD
          valueFrom:
            secretKeyRef:
              name: xql-zero-xql-secret
              key: password
        - name: XQL_ONE_ARG
          valueFrom:
            secretKeyRef:
              name: $(PROJECT_XQL_ONE_SECRET_NAME)
              key: arg
        - name: XQL_ONE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: xql-one-xql-secret
              key: password
`)

	th.WriteK("bases/xql", `
secretGenerator:
- name: xql-secret
  envs:
  - config/xql-secret.env
`)

	th.WriteF("bases/xql/config/xql-secret.env", `
password=SUPER_SECRET_PASSWORD
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: project-app
spec:
  template:
    spec:
      containers:
      - env:
        - name: XQL_ZERO_ARG
          valueFrom:
            secretKeyRef:
              key: arg
              name: project-xql-zero-xql-secret-6khmtc56hm
        - name: XQL_ZERO_PASSWORD
          valueFrom:
            secretKeyRef:
              key: password
              name: project-xql-zero-xql-secret-6khmtc56hm
        - name: XQL_ONE_ARG
          valueFrom:
            secretKeyRef:
              key: arg
              name: project-xql-one-xql-secret-79mhmf5dgt
        - name: XQL_ONE_PASSWORD
          valueFrom:
            secretKeyRef:
              key: password
              name: project-xql-one-xql-secret-79mhmf5dgt
        image: example.com/app:latest
        imagePullPolicy: Always
        name: app
---
apiVersion: v1
data:
  arg: MA==
  password: U1VQRVJfU0VDUkVUX1BBU1NXT1JE
kind: Secret
metadata:
  name: project-xql-zero-xql-secret-6khmtc56hm
type: Opaque
---
apiVersion: v1
data:
  arg: MQ==
  password: U1VQRVJfU0VDUkVUX1BBU1NXT1JE
kind: Secret
metadata:
  name: project-xql-one-xql-secret-79mhmf5dgt
type: Opaque
`)
}
