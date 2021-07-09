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
	th.WriteF("replacement.yaml", `
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

	th.WriteF("base/deployments.yaml", `
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
	th.WriteK("base", `
resources:
- deployments.yaml
`)
	th.WriteK("a", `
namePrefix: a-
resources:
- ../base
replacements:
- path: replacement.yaml
`)
	th.WriteF("a/replacement.yaml", `
source:
  name: a-sourceA
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    name: a-deploy
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)
	th.WriteK("b", `
namePrefix: b-
resources:
- ../base
replacements:
- path: replacement.yaml
`)
	th.WriteF("b/replacement.yaml", `
source:
  name: b-sourceB
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    name: b-deploy
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)
	th.WriteK(".", `
resources:
- a
- b
`)

	m := th.Run(".", th.MakeDefaultOptions())
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

func TestReplacementTransformerWithOriginalName(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/deployments.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: target
spec:
  template:
    spec:
      containers:
      - image: nginx:oldtag
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: source
spec:
  template:
    spec:
      containers:
      - image: nginx:newtag
        name: nginx
`)
	th.WriteK("base", `
resources:
- deployments.yaml
`)
	th.WriteK("overlay", `
namePrefix: prefix1-
resources:
- ../base
`)

	th.WriteK(".", `
namePrefix: prefix2-
resources:
- overlay
replacements:
- path: replacement.yaml
`)
	th.WriteF("replacement.yaml", `
source:
  name: source
  fieldPath: spec.template.spec.containers.0.image
targets:
- select:
    name: prefix1-target
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix2-prefix1-target
spec:
  template:
    spec:
      containers:
      - image: nginx:newtag
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix2-prefix1-source
spec:
  template:
    spec:
      containers:
      - image: nginx:newtag
        name: nginx
`)
}

// TODO: Address namePrefix in overlay not applying to replacement targets
// The image in the target deployment should end up being `prefix-source` instead of `source`
// https://github.com/kubernetes-sigs/kustomize/issues/4034
func TestReplacementTransformerWithNamePrefixOverlay(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/deployments.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: target
spec:
  template:
    spec:
      containers:
      - image: nginx:oldtag
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: source
spec:
  template:
    spec:
      containers:
      - image: nginx:newtag
        name: nginx
`)
	th.WriteF("base/replacement.yaml", `
source:
  name: source
  kind: Deployment
targets:
- select:
    name: target
  fieldPaths:
  - spec.template.spec.containers.[name=nginx].image
`)
	th.WriteK("base", `
replacements:
- path: replacement.yaml
resources:
- deployments.yaml
`)

	th.WriteK(".", `
namePrefix: prefix-
resources:
- base
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix-target
spec:
  template:
    spec:
      containers:
      - image: source
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prefix-source
spec:
  template:
    spec:
      containers:
      - image: nginx:newtag
        name: nginx
`)
}
