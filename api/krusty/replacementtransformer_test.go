// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

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

func TestReplacementsFieldWithPathMultiple(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
resources:
- resource.yaml

replacements:
- path: replacement.yaml
`)
	th.WriteF("replacement.yaml", `
- source: 
    kind: Deployment
    fieldPath: spec.template.spec.containers.0.image
  targets:
  - select:
      kind: Deployment
    fieldPaths: 
    - spec.template.spec.containers.1.image
- source: 
    kind: Deployment
    fieldPath: spec.template.spec.containers.0.name
  targets:
  - select:
      kind: Deployment
    fieldPaths: 
    - spec.template.spec.containers.1.name
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
        name: replaced-with-digest
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
// The property `data.blue-name` should end up being `overlay-blue` instead of `blue`
// https://github.com/kubernetes-sigs/kustomize/issues/4034
func TestReplacementTransformerWithNamePrefixOverlay(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("base", `
generatorOptions:
 disableNameSuffixHash: true
configMapGenerator:
- name: blue
- name: red
replacements:
- source:
    kind: ConfigMap
    name: blue
    fieldPath: metadata.name
  targets:
  - select:
      name: red
    fieldPaths:
    - data.blue-name
    options:
      create: true
`)

	th.WriteK(".", `
namePrefix: overlay-
resources:
- base
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: overlay-blue
---
apiVersion: v1
data:
  blue-name: blue
kind: ConfigMap
metadata:
  name: overlay-red
`)
}

// TODO: Address namespace in overlay not applying to replacement targets
// The property `data.blue-namespace` should end up being `overlay-namespace` instead of `base-namespace`
// https://github.com/kubernetes-sigs/kustomize/issues/4034
func TestReplacementTransformerWithNamespaceOverlay(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK("base", `
namespace: base-namespace
generatorOptions:
 disableNameSuffixHash: true
configMapGenerator:
- name: blue
- name: red
replacements:
- source:
    kind: ConfigMap
    name: blue
    fieldPath: metadata.namespace
  targets:
  - select:
      name: red
    fieldPaths:
    - data.blue-namespace
    options:
      create: true
`)

	th.WriteK(".", `
namespace: overlay-namespace
resources:
- base
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: blue
  namespace: overlay-namespace
---
apiVersion: v1
data:
  blue-namespace: base-namespace
kind: ConfigMap
metadata:
  name: red
  namespace: overlay-namespace
`)
}

// TODO: Address configMapGenerator suffix not applying to replacement targets
// The property `data.blue-name` should end up being `blue-6ct58987ht` instead of `blue`
// https://github.com/kubernetes-sigs/kustomize/issues/4034
func TestReplacementTransformerWithConfigMapGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
configMapGenerator:
- name: blue
- name: red
replacements:
- source:
    kind: ConfigMap
    name: blue
    fieldPath: metadata.name
  targets:
  - select:
      name: red
    fieldPaths:
    - data.blue-name
    options:
      create: true
`)

	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: ConfigMap
metadata:
  name: blue-6ct58987ht
---
apiVersion: v1
data:
  blue-name: blue
kind: ConfigMap
metadata:
  name: red-dc6gc5btkc
`)
}

func TestReplacementTransformerWithSuffixTransformerAndReject(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/app.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: original-name
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
`)
	th.WriteK("base", `
resources:
  - app.yaml
`)
	th.WriteK("overlay", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

nameSuffix: -dev
resources:
  - ../base

configMapGenerator:
  - name: app-config
    literals:
      - name=something-else

replacements:
  - source:
      kind: ConfigMap
      name: app-config
      fieldPath: data.name
    targets:
      - fieldPaths:
          - spec.template.spec.containers.0.name
        select:
          kind: Deployment
        reject:
          - name: original-name
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: original-name-dev
spec:
  template:
    spec:
      containers:
      - image: app1:1.0
        name: app
---
apiVersion: v1
data:
  name: something-else
kind: ConfigMap
metadata:
  name: app-config-dev-97544dk6t8
`)
}

// regex selector: append in annotation by visitor name
func TestReplacementTransformerAppendToAnnotationUsingRegex(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/app1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
`)
	th.WriteF("base/app2.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d2
spec:
  template:
    spec:
      containers:
        - image: app2:1.0
          name: app
`)
	th.WriteF("base/cm1.yaml", `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: cm1
`)
	th.WriteF("base/cm2.yaml", `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: cm2
`)
	th.WriteF("base/pg1.yaml", `
apiVersion: apps/v1
kind: postgresql
metadata:
  name: pg1
`)
	th.WriteK("base", `
resources:
- app1.yaml
- app2.yaml
- cm1.yaml
- cm2.yaml
- pg1.yaml

replacements:
  - source:
      kind: ConfigMap
      name: cm1
    targets:
      - reject:
          - kind: ConfigMap
            name: c.1
        select:
          kind: Deployment|ConfigMap|postgresql
        fieldPaths:
          - metadata.annotations.visitedby
        options:
          index: -1
          delimiter: ","
          create: true
  - source:
      kind: ConfigMap
      name: cm2
    targets:
      - reject:
          - kind: ConfigMap
            name: .*2
        select:
          kind: Deployment|ConfigMap|postgresql
        fieldPaths:
          - metadata.annotations.visitedby
        options:
          index: -1
          delimiter: ","
          create: true
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    visitedby: cm2,cm1,
  name: d1
spec:
  template:
    spec:
      containers:
      - image: app1:1.0
        name: app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    visitedby: cm2,cm1,
  name: d2
spec:
  template:
    spec:
      containers:
      - image: app2:1.0
        name: app
---
apiVersion: apps/v1
kind: ConfigMap
metadata:
  annotations:
    visitedby: cm2,
  name: cm1
---
apiVersion: apps/v1
kind: ConfigMap
metadata:
  annotations:
    visitedby: cm1,
  name: cm2
---
apiVersion: apps/v1
kind: postgresql
metadata:
  annotations:
    visitedby: cm2,cm1,
  name: pg1
`)
}

// selector regex: construct service url
func TestReplacementTransformerServiceNamespaceUrlUsingRegex(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/d1.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
          env:
            - name: APP1_SERVICE
              value: "d1.app1"
`)
	th.WriteF("base/d2.yaml", `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: d2
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
          env:
            - name: APP1_SERVICE
              value: "d2.app1"
`)
	th.WriteF("base/sts1.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: sts1
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
          env:
            - name: APP1_SERVICE
              value: "app1"
`)
	th.WriteF("base/cm1.yaml", `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: cm1
data:
  APP1_SERVICE_PORT: "8080"
`)
	th.WriteF("base/svc1.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: svc1-namespace
spec:
  selector:
    app.kubernetes.io/name: app1
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376

`)
	th.WriteK("base", `
resources:
- d1.yaml
- d2.yaml
- sts1.yaml
- cm1.yaml
- svc1.yaml

replacements:
  - source:
      kind: Service
      name: svc1
      fieldPath: metadata.namespace
    targets:
      - select:
          kind: Deployment|.*Set
        fieldPaths:
          - spec.template.spec.containers.*.env.[name=APP1_SERVICE].value
        options:
          index: 99
          delimiter: "."
  - source:
      kind: ConfigMap
      name: cm1
      fieldPath: data.APP1_SERVICE_PORT
    targets:
      - select:
          kind: Deployment|.*Set
        fieldPaths:
          - spec.template.spec.containers.*.env.[name=APP1_SERVICE].value
        options:
          index: 99
          delimiter: ":"
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
spec:
  template:
    spec:
      containers:
      - env:
        - name: APP1_SERVICE
          value: d1.app1.svc1-namespace:8080
        image: app1:1.0
        name: app
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: d2
spec:
  template:
    spec:
      containers:
      - env:
        - name: APP1_SERVICE
          value: d2.app1.svc1-namespace:8080
        image: app1:1.0
        name: app
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: sts1
spec:
  template:
    spec:
      containers:
      - env:
        - name: APP1_SERVICE
          value: app1.svc1-namespace:8080
        image: app1:1.0
        name: app
---
apiVersion: apps/v1
data:
  APP1_SERVICE_PORT: "8080"
kind: ConfigMap
metadata:
  name: cm1
---
apiVersion: v1
kind: Service
metadata:
  name: svc1
  namespace: svc1-namespace
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 9376
  selector:
    app.kubernetes.io/name: app1
`)
}

func TestReplacementTransformerWithSuffixTransformerAndRejectUsingRegex(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("base/app.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: original-name
spec:
  template:
    spec:
      containers:
        - image: app1:1.0
          name: app
`)
	th.WriteK("base", `
resources:
  - app.yaml
`)
	th.WriteK("overlay", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

nameSuffix: -dev
namePrefix: pre-
resources:
  - ../base

configMapGenerator:
  - name: app-config
    literals:
      - name=something-else

replacements:
- source:
    kind: ConfigMap
    name: app-config
    fieldPath: data.name
  targets:
    - reject:
        - name: .*original.*
      select:
        kind: Deployment
      fieldPaths:
        - spec.template.spec.containers.0.name
    - select:
        kind: ConfigMap
        name: app-config
      fieldPaths:
        - data.name-copy
      options:
        create: true
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pre-original-name-dev
spec:
  template:
    spec:
      containers:
      - image: app1:1.0
        name: app
---
apiVersion: v1
data:
  name: something-else
  name-copy: something-else
kind: ConfigMap
metadata:
  name: pre-app-config-dev-7266b7f2m9
`)
}
