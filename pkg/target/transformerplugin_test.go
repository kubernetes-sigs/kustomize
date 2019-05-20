// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/plugin"
)

func writeDeployment(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
}

func writeStringPrefixer(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: StringPrefixer
metadata:
  name: apple
`)
}

func writeDatePrefixer(th *kusttest_test.KustTestHarness, path string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: irrelevant
`)
}

func TestOrderedTransformers(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	th.WriteK("/app", `
resources:
- deployment.yaml
transformers:
- stringPrefixer.yaml
# - datePrefixer.yaml
`)
	// TODO(monopole): assure ordering of loaded
	// transformers and this will work - the trouble
	// is we load into a map (ResMap), not a list.
	writeDeployment(th, "/app/deployment.yaml")
	writeStringPrefixer(th, "/app/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/datePrefixer.yaml")
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apple-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}

func TestSedTransformer(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "SedTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")
	th.WriteK("/app", `
resources:
- configmap.yaml

transformers:
- sed-transformer.yaml

configMapGenerator:
- name: test
  literals:
  - FOO=$FOO
  - BAR=$BAR
`)
	th.WriteF("/app/sed-transformer.yaml", `
apiVersion: someteam.example.com/v1
kind: SedTransformer
metadata:
  name: some-random-name
argsFromFile: sed-input.txt
`)
	th.WriteF("/app/sed-input.txt", `
s/$FOO/foo/g
s/$BAR/bar/g
`)

	th.WriteF("/app/configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-a
  annotations:
    kustomize.k8s.io/Generated: "false"
data:
  foo: $FOO
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  foo: foo
kind: ConfigMap
metadata:
  annotations:
    kustomize.k8s.io/Generated: "false"
  name: configmap-a
---
apiVersion: v1
data:
  BAR: bar
  FOO: foo
kind: ConfigMap
metadata:
  annotations: {}
  name: test-k4bkhftttd
`)
}

func TestTransformedTransformers(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "StringPrefixer")

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "DatePrefixer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app/overlay")

	th.WriteK("/app/base", `
resources:
- stringPrefixer.yaml
transformers:
- datePrefixer.yaml
`)
	writeStringPrefixer(th, "/app/base/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/base/datePrefixer.yaml")

	th.WriteK("/app/overlay", `
resources:
- deployment.yaml
transformers:
- ../base
`)
	writeDeployment(th, "/app/overlay/deployment.yaml")

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 2018-05-11-apple-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}
