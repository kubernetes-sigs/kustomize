// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

func writeStringPrefixer(th *kusttest_test.HarnessEnhanced, path, name string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: StringPrefixer
metadata:
  name: `+name+`
`)
}

func TestPluginsNotEnabled(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "StringPrefixer")
	defer th.Reset()

	th.WriteK("/app", `
transformers:
- stringPrefixer.yaml
`)
	writeStringPrefixer(th, "/app/stringPrefixer.yaml", "apple")
	err := th.RunWithErr("/app", th.MakeOptionsPluginsDisabled())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !types.IsErrOnlyBuiltinPluginsAllowed(err) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestSedTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("someteam.example.com", "v1", "SedTransformer")
	defer th.Reset()

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

	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
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
  name: test-6bc28fff49
`)
}

/*

The tests below are disabled until the StringPrefixer and DatePrefixer
can be rewritten using kyaml, instead of depending on the
PrefixSuffixTransformerPlugin.  That dependency is causing
failures in the test loader.

func writeDeployment(th *kusttest_test.HarnessEnhanced, path string) {
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

func writeDatePrefixer(th *kusttest_test.HarnessEnhanced, path, name string) {
	th.WriteF(path, `
apiVersion: someteam.example.com/v1
kind: DatePrefixer
metadata:
  name: `+name+`
`)
}

func TestOrderedTransformers(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "StringPrefixer").
		BuildGoPlugin("someteam.example.com", "v1", "DatePrefixer")
	defer th.Reset()

	th.WriteK("/app", `
resources:
- deployment.yaml
transformers:
- peachPrefixer.yaml
- date1Prefixer.yaml
- applePrefixer.yaml
- date2Prefixer.yaml
`)
	writeDeployment(th, "/app/deployment.yaml")
	writeStringPrefixer(th, "/app/applePrefixer.yaml", "apple")
	writeStringPrefixer(th, "/app/peachPrefixer.yaml", "peach")
	writeDatePrefixer(th, "/app/date1Prefixer.yaml", "date1")
	writeDatePrefixer(th, "/app/date2Prefixer.yaml", "date2")
	m := th.Run("/app", th.MakeOptionsPluginsEnabled())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 2018-05-11-apple-2018-05-11-peach-myDeployment
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

func TestTransformedTransformers(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("someteam.example.com", "v1", "StringPrefixer").
		BuildGoPlugin("someteam.example.com", "v1", "DatePrefixer")
	defer th.Reset()

	th.WriteK("/app/base", `
resources:
- stringPrefixer.yaml
transformers:
- datePrefixer.yaml
`)
	writeStringPrefixer(th, "/app/base/stringPrefixer.yaml", "apple")
	writeDatePrefixer(th, "/app/base/datePrefixer.yaml", "date")

	th.WriteK("/app/overlay", `
resources:
- deployment.yaml
transformers:
- ../base
`)
	writeDeployment(th, "/app/overlay/deployment.yaml")

	m := th.Run("/app/overlay", th.MakeOptionsPluginsEnabled())
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
*/
