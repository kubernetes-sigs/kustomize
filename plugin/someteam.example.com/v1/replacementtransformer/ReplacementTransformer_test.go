// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestReplacementTransformer(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "ReplacementTransformer")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image
- source:
    value: postgres:latest
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.3.image
`, `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:newtag
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:latest
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`)
}

func TestReplacementTransformerComplexType(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "ReplacementTransformer")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: Pod
      name: pod
    fieldref: spec.containers
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers
`, `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
`)
}

func TestReplacementTransformerFromConfigMap(t *testing.T) {
	tc := kusttest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "ReplacementTransformer")

	th := kusttest_test.NewKustTestHarnessAllowPlugins(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: someteam.example.com/v1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.HOSTNAME
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[image=debian].args.0
    - spec.template.spec.containers[name=busybox].args.1
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.PORT
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[image=debian].args.1
    - spec.template.spec.containers[name=busybox].args.2
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: command-demo-container
          image: debian
          command: ["printenv"]
          args:
            - HOSTNAME
            - PORT
        - name: busybox
          image: busybox:latest
          args:
            - echo
            - HOSTNAME
            - PORT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
      - args:
        - example.com
        - 8080
        command:
        - printenv
        image: debian
        name: command-demo-container
      - args:
        - echo
        - example.com
        - 8080
        image: busybox:latest
        name: busybox
---
apiVersion: v1
data:
  HOSTNAME: example.com
  PORT: 8080
kind: ConfigMap
metadata:
  name: cm
`)
}