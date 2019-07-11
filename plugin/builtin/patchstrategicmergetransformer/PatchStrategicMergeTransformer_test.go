// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
)

const (
	target = `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
	targetWithNamespace = `
apiVersion: apps/v1
metadata:
  name: myDeploy
  namespace: namespace1
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
	targetNoschema = `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    B: Y
`
)

func TestPatchStrategicMergeTransformerMissingFile(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")
	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(),
		"cannot read file \"/app/patch.yaml\"") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBadPatchStrategicMergeTransformer(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
patches: 'thisIsNotAPatch'
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(),
		"cannot unmarshal string into Go value of type map[string]interface {}") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBothEmptyPatchStrategicMergeTransformer(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	_, err := th.RunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
`, target)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "empty file path and empty patch content") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPatchStrategicMergeTransformerFromFiles(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 3
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, target)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchStrategicMergeTransformerWithInline(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
patches: '{"apiVersion": "apps/v1", "metadata": {"name": "myDeploy"}, "kind": "Deployment", "spec": {"replica": 3}}'
`, target)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchStrategicMergeTransformerMultiplePatches(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch1.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        env:
        - name: SOMEENV
          value: BAR
`)

	th.WriteF("/app/patch2.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ANOTHERENV
          value: HELLO
      - name: busybox
        image: busybox
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, target)

	th.AssertActualEqualsExpected(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - env:
        - name: ANOTHERENV
          value: HELLO
        - name: SOMEENV
          value: BAR
        image: nginx:latest
        name: nginx
      - image: busybox
        name: busybox
`)
}

func TestStrategicMergeTransformerMultiplePatchesWithConflicts(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch1.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        env:
        - name: SOMEENV
          value: BAR
`)

	th.WriteF("/app/patch2.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        env:
        - name: ANOTHERENV
          value: HELLO
      - name: busybox
        image: busybox
`)

	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, target)

	if err == nil {
		t.Fatalf("did not get expected error")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("expected error to contain %q but get %v", "conflict", err)
	}
}

func TestStrategicMergeTransformerWrongNamespace(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
  namespace: namespace2
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        env:
        - name: SOMEENV
          value: BAR
`)

	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, targetWithNamespace)

	if err == nil {
		t.Fatalf("did not get expected error")
	}
	if !strings.Contains(err.Error(), "failed to find unique target for patch") {
		t.Fatalf("expected error to contain %q but get %v", "failed to find target for patch", err)
	}
}

func TestStrategicMergeTransformerNoSchema(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, targetNoschema)

	th.AssertActualEqualsExpected(rm, `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    C: Z
`)
}

func TestStrategicMergeTransformerNoSchemaMultiPatches(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch1.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`)
	th.WriteF("/app/patch2.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: Z
    D: W
  baz:
    hello: world
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, targetNoschema)

	th.AssertActualEqualsExpected(rm, `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    C: Z
    D: W
  baz:
    hello: world
`)
}

func TestStrategicMergeTransformerNoSchemaMultiPatchesWithConflict(t *testing.T) {
	tc := plugins.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchStrategicMergeTransformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	th.WriteF("/app/patch1.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: Z
`)
	th.WriteF("/app/patch2.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: NOT_Z

`)
	err := th.ErrorFromLoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, targetNoschema)
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("expected error to contain %q but get %v", "conflict", err)
	}
}
