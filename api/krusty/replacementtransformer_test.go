package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// end to end tests to demonstrate functionality of ReplacementTransformer
func TestReplacementsField(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
resources:
- resource.yaml

replacements:
- source:
    kind: Deployment
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    fieldPaths:
    - spec.template.spec.containers.1.image
`)
	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
`)
	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: foobar:1
        name: postgresdb
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestReplacementsFieldWithPath(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
resources:
- resource.yaml

replacements:
- path: replacement.yaml
`)
	th.WriteF("replacement.yaml",
		`
source: 
  kind: Deployment
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    kind: Deployment
  fieldPaths: 
  - spec.template.spec.containers.1.image
`)
	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
`)
	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: foobar:1
        name: replaced-with-digest
      - image: foobar:1
        name: postgresdb
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestReplacementTransformerWithDiamondShape(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/deployment.yaml",
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sourceA
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagA
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sourceB
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagB
        name: nginx
`)
	th.WriteF("base/kustomization.yaml",
		`
resources:
- deployment.yaml
`)
	th.WriteF("a/kustomization.yaml",
		`
namePrefix: a-
resources:
- ../base
replacements:
- path: replacement.yaml
`)
	th.WriteF("a/replacement.yaml",
		`
source:
  name: a-sourceA
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    name: a-deploy
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)
	th.WriteF("b/kustomization.yaml",
		`
namePrefix: b-
resources:
- ../base
replacements:
- path: replacement.yaml
`)
	th.WriteF("b/replacement.yaml",
		`
source:
  name: b-sourceB
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    name: b-deploy
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)
	th.WriteF("combined/kustomization.yaml",
		`
resources:
- ../a
- ../b
`)

	m := th.Run("combined", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a-deploy
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagA
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a-sourceA
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagA
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a-sourceB
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagB
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-deploy
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagB
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-sourceA
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagA
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b-sourceB
spec:
  template:
    spec:
      containers:
      - image: nginx:newtagB
        name: nginx
`)
}
