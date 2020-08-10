// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"fmt"
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"'/patch.yaml' doesn't exist") &&
			!strings.Contains(err.Error(),
				"cannot unmarshal string") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBadPatchStrategicMergeTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
patches: 'thisIsNotAPatch'
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"cannot unmarshal string into Go value of type map[string]interface {}") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestBothEmptyPatchStrategicMergeTransformer(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "empty file path and empty patch content") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchStrategicMergeTransformerFromFiles(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("/patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  template:
    metadata:
      labels:
        new-label: new-value
  replica: 3
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`,
		target,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        new-label: new-value
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchStrategicMergeTransformerWithInlineJSON(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
patches: '{"apiVersion": "apps/v1", "metadata": {"name": "myDeploy"}, "kind": "Deployment", "spec": {"replica": 3}}'
`,
		target,
		`
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

func TestPatchStrategicMergeTransformerWithInlineYAML(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
patches: |-
  apiVersion: apps/v1
  metadata:
    name: myDeploy
  kind: Deployment
  spec:
    replica: 3
  ---
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
`,
		target,
		`
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
      - image: nginx:latest
        name: nginx
`)
}

func TestPatchStrategicMergeTransformerMultiplePatches(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch1.yaml", `
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

	th.WriteF("patch2.yaml", `
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

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`,
		target,
		`
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch1.yaml", `
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

	th.WriteF("patch2.yaml", `
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

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, target, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("did not get expected error")
		}
		if !strings.Contains(err.Error(), "conflict") {
			t.Fatalf("expected error to contain %q but get %v", "conflict", err)
		}
	})
}

func TestStrategicMergeTransformerWrongNamespace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
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

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, targetWithNamespace, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("did not get expected error")
		}
		if !strings.Contains(err.Error(), "failed to find unique target for patch") {
			t.Fatalf("expected error to contain %q but get %v", "failed to find target for patch", err)
		}
	})
}

// issue #2734 -- https://github.com/kubernetes-sigs/kustomize/issues/2734
// kyaml cleans up things some folks prefer it didnt
// 1. []    -- see initContainers and imagePullSecrets
// 2. {}    -- see emptyDir
// 3. null  -- see creationTimestamp

const anUncleanDeploymentResource = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sas-crunchy-data-postgres-operator
spec:
  replicas: 1
  template:
    metadata:
      creationTimestamp: null
    spec:
      serviceAccountName: postgres-operator
      containers:
      - name: apiserver
        image: sas-crunchy-data-operator-api-server
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8443
        envFrom: []
        volumeMounts:
        - name: security-ssh
          mountPath: /security-ssh
        - mountPath: /tmp
          name: tmp
      imagePullSecrets: []
      initContainers: []
      volumes:
      - emptyDir: {}
        name: security-ssh
      - emptyDir: {}
        name: tmp
`

// This is the preffered result (and what we get with kustomize 3.7.0)
const expectedCleanedDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    workload.sas.com/class: stateless
  name: sas-crunchy-data-postgres-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        workload.sas.com/class: stateless
    spec:
      containers:
      - envFrom: []
        image: sas-crunchy-data-operator-api-server
        imagePullPolicy: IfNotPresent
        name: apiserver
        ports:
        - containerPort: 8443
        volumeMounts:
        - mountPath: /security-ssh
          name: security-ssh
        - mountPath: /tmp
          name: tmp
      imagePullSecrets: []
      initContainers: []
      serviceAccountName: postgres-operator
      tolerations:
      - effect: NoSchedule
        key: workload.sas.com/class
        operator: Equal
        value: stateful
      volumes:
      - emptyDir: {}
        name: security-ssh
      - emptyDir: {}
        name: tmp
`

func TestPatchStrategicMergeTransformerCleanupItems(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("/patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sas-crunchy-data-postgres-operator
  labels:
    workload.sas.com/class: stateless
spec:
  template:
    metadata:
      labels:
        workload.sas.com/class: stateless
    spec:
      tolerations:
        - effect: NoSchedule
          key: workload.sas.com/class
          operator: Equal
          value: stateful
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`,
		anUncleanDeploymentResource,
		expectedCleanedDeployment) // prefer expectedCleanedDeployment
}

func TestStrategicMergeTransformerNoSchema(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`)
	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`,
		targetNoschema,
		`
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch1.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`)
	th.WriteF("patch2.yaml", `
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
	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`,
		targetNoschema,
		`
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch1.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: Z
`)
	th.WriteF("patch2.yaml", `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: NOT_Z

`)
	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch1.yaml
- patch2.yaml
`, targetNoschema, func(t *testing.T, err error) {
		if !strings.Contains(err.Error(), "conflict") {
			t.Fatalf("expected error to contain %q but get %v", "conflict", err)
		}
	})

}

// simple utility function to add an namespace in a resource
// used as base, patch or expected result. Simply looks
// for specs: in order to add namespace: xxxx before this line
func addNamespace(namespace string, base string) string {
	res := strings.Replace(base,
		"\nspec:\n",
		"\n  namespace: "+namespace+"\nspec:\n",
		1)
	return res
}

// compareExpectedError compares the expectedError and the actualError return by GetFieldValue
func compareExpectedError(t *testing.T, name string, err error, errorMsg string) {
	if err == nil {
		t.Fatalf("%q; - should return error, but no error returned", name)
	}

	if !strings.Contains(err.Error(), errorMsg) {
		t.Fatalf("%q; - expected error: \"%s\", got error: \"%v\"",
			name, errorMsg, err.Error())
	}
}

const Deployment string = "Deployment"
const MyCRD string = "MyCRD"

// baseResource produces a base object which used to test
// patch transformation
// Also the structure is matching the Deployment syntax
// the kind can be replaced to allow testing using CRD
// without access to the schema
func baseResource(kind string) string {

	res := `
apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx`
	return fmt.Sprintf(res, kind)
}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a label in the metadata
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func addLabelAndEnvPatch(kind string) string {

	res := `
apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        some-label: some-value
    spec:
      containers:
       - name: nginx
         env:
         - name: SOMEENV
           value: SOMEVALUE`

	return fmt.Sprintf(res, kind)
}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a second container in the container list
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func addContainerAndEnvPatch(kind string) string {

	res := `
apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: ANOTHERENV
          value: ANOTHERVALUE
      - name: anothercontainer
        image: anotherimage`

	return fmt.Sprintf(res, kind)
}

// addContainerAndEnvPatch produces a patch object which replaces
// the value of the image field in the first/nginx container
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func changeImagePatch(kind string, newImage string) string {

	res := `
apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: %s`

	return fmt.Sprintf(res, kind, newImage)
}

// utility method building the expected output of a SMP
func expectedResultSMP() string {

	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: SOMEENV
          value: SOMEVALUE
        image: nginx
        name: nginx
`
}

// utility method building the expected output of a JMP.
// imagename parameter allows to build a result consistent
// with the JMP behavior which basically overrides the
// entire "containers" list.
func expectedResultJMP(imagename string) string {

	res := `apiVersion: apps/v1
kind: MyCRD
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: SOMEENV
          value: SOMEVALUE
        name: nginx
`

	if imagename == "" {
		return res
	}

	res = `apiVersion: apps/v1
kind: MyCRD
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - image: %s
        name: nginx
`

	return fmt.Sprintf(res, imagename)
}

// utility method to build the expected result of a multipatch
// the order of the patches still have influence especially
// in the insertion location within arrays.
func expectedResultMultiPatch(kind string, reversed bool) string {

	res := `apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: ANOTHERENV
          value: ANOTHERVALUE
        - name: SOMEENV
          value: SOMEVALUE
        image: nginx:latest
        name: nginx
      - image: anotherimage
        name: anothercontainer
`

	reversedres := `apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    metadata:
      labels:
        old-label: old-value
        some-label: some-value
    spec:
      containers:
      - env:
        - name: SOMEENV
          value: SOMEVALUE
        - name: ANOTHERENV
          value: ANOTHERVALUE
        image: nginx:latest
        name: nginx
      - image: anotherimage
        name: anothercontainer
`

	if reversed {
		return fmt.Sprintf(reversedres, kind)
	}

	return fmt.Sprintf(res, kind)
}

func toConfig(patches ...string) string {
	config := `
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
`
	for idx := range patches {
		config = fmt.Sprintf("%s\n- ./patch%d.yaml", config, idx)
	}

	return config
}

// TestSinglePatch validates the single patch use cases
// regarless of the schema availibility, which in turns
// relies on StrategicMergePatch or simple JSON Patch.
func TestSinglePatch(t *testing.T) {
	tests := []struct {
		name          string
		base          string
		patch         string
		expected      string
		errorExpected bool
		errorMsg      string
	}{
		{
			name:          "withschema",
			base:          baseResource(Deployment),
			patch:         addLabelAndEnvPatch(Deployment),
			errorExpected: false,
			expected:      expectedResultSMP(),
		},
		{
			name:          "noschema",
			base:          baseResource(MyCRD),
			patch:         addLabelAndEnvPatch(MyCRD),
			errorExpected: false,
			expected:      expectedResultJMP(""),
		},
	}

	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	for _, test := range tests {
		th.ResetLoaderRoot(fmt.Sprintf("/%s", test.name))
		th.WriteF(fmt.Sprintf("/%s/patch%d.yaml", test.name, 0), test.patch)
		if test.errorExpected {
			th.RunTransformerAndCheckError(toConfig(test.patch), test.base,
				func(t *testing.T, err error) {
					compareExpectedError(t, test.name, err, test.errorMsg)
				})
		} else {
			th.RunTransformerAndCheckResult(toConfig(test.patch), test.base,
				test.expected)
		}
	}
}

// TestMultiplePatches checks that the patches are applied
// properly, that the same result is obtained,
// regardless of the order of the patches and regardless
// of the schema availibility (SMP vs JSON)
func TestMultiplePatches(t *testing.T) {
	tests := []struct {
		name          string
		base          string
		patch         []string
		expected      string
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-label-image-container",
			base: baseResource(Deployment),
			patch: []string{
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:latest"),
				addContainerAndEnvPatch(Deployment),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(Deployment, false),
		},
		{
			name: "withschema-image-container-label",
			base: baseResource(Deployment),
			patch: []string{
				changeImagePatch(Deployment, "nginx:latest"),
				addContainerAndEnvPatch(Deployment),
				addLabelAndEnvPatch(Deployment),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(Deployment, true),
		},
		{
			name: "withschema-container-label-image",
			base: baseResource(Deployment),
			patch: []string{
				addContainerAndEnvPatch(Deployment),
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:latest"),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(Deployment, true),
		},
		{
			name: "noschema-label-image-container",
			base: baseResource(MyCRD),
			patch: []string{
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:latest"),
				addContainerAndEnvPatch(MyCRD),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-image-container-label",
			base: baseResource(MyCRD),
			patch: []string{
				changeImagePatch(MyCRD, "nginx:latest"),
				addContainerAndEnvPatch(MyCRD),
				addLabelAndEnvPatch(MyCRD),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-container-label-image",
			base: baseResource(MyCRD),
			patch: []string{
				addContainerAndEnvPatch(MyCRD),
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:latest"),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
	}

	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	for _, test := range tests {
		th.ResetLoaderRoot(fmt.Sprintf("/%s", test.name))
		for idx, patch := range test.patch {
			th.WriteF(fmt.Sprintf("/%s/patch%d.yaml", test.name, idx), patch)
		}

		if test.errorExpected {
			th.RunTransformerAndCheckError(toConfig(test.patch...), test.base,
				func(t *testing.T, err error) {
					compareExpectedError(t, test.name, err, test.errorMsg)
				})
		} else {
			th.RunTransformerAndCheckResult(toConfig(test.patch...), test.base,
				test.expected)
		}
	}

}

// TestMultiplePatchesWithConflict checks that the conflict are
// detected regardless of the order of the patches and regardless
// of the schema availibility (SMP vs JSON)
func TestMultiplePatchesWithConflict(t *testing.T) {
	tests := []struct {
		name          string
		base          string
		patch         []string
		expected      string
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-label-latest-1.7.9",
			base: baseResource(Deployment),
			patch: []string{
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:latest"),
				changeImagePatch(Deployment, "nginx:1.7.9"),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-latest-label-1.7.9",
			base: baseResource(Deployment),
			patch: []string{
				changeImagePatch(Deployment, "nginx:latest"),
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:1.7.9"),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-1.7.9-label-latest",
			base: baseResource(Deployment),
			patch: []string{
				changeImagePatch(Deployment, "nginx:1.7.9"),
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:latest"),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-1.7.9-latest-label",
			base: baseResource(Deployment),
			patch: []string{
				changeImagePatch(Deployment, "nginx:1.7.9"),
				changeImagePatch(Deployment, "nginx:latest"),
				addLabelAndEnvPatch(Deployment),
				changeImagePatch(Deployment, "nginx:nginx"),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-label-latest-1.7.9",
			base: baseResource(MyCRD),
			patch: []string{
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:latest"),
				changeImagePatch(MyCRD, "nginx:1.7.9"),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-latest-label-1.7.9",
			base: baseResource(MyCRD),
			patch: []string{
				changeImagePatch(MyCRD, "nginx:latest"),
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:1.7.9"),
			},
			errorExpected: false,
			// There is no conflict detected. It should
			// be but the JMPConflictDector ignores it.
			// See https://github.com/kubernetes-sigs/kustomize/issues/1370
			expected: expectedResultJMP("nginx:1.7.9"),
		},
		{
			name: "noschema-1.7.9-label-latest",
			base: baseResource(MyCRD),
			patch: []string{
				changeImagePatch(MyCRD, "nginx:1.7.9"),
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:latest"),
			},
			errorExpected: false,
			// There is no conflict detected. It should
			// be but the JMPConflictDector ignores it.
			// See https://github.com/kubernetes-sigs/kustomize/issues/1370
			expected: expectedResultJMP("nginx:latest"),
		},
		{
			name: "noschema-1.7.9-latest-label",
			base: baseResource(MyCRD),
			patch: []string{
				changeImagePatch(MyCRD, "nginx:1.7.9"),
				changeImagePatch(MyCRD, "nginx:latest"),
				addLabelAndEnvPatch(MyCRD),
				changeImagePatch(MyCRD, "nginx:nginx"),
			},
			errorExpected: true,
		},
	}

	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	for _, test := range tests {
		th.ResetLoaderRoot(fmt.Sprintf("/%s", test.name))
		for idx, patch := range test.patch {
			th.WriteF(fmt.Sprintf("/%s/patch%d.yaml", test.name, idx), patch)
		}

		if test.errorExpected {
			th.RunTransformerAndCheckError(toConfig(test.patch...), test.base,
				func(t *testing.T, err error) {
					compareExpectedError(t, test.name, err, test.errorMsg)
				})
		} else {
			th.RunTransformerAndCheckResult(toConfig(test.patch...), test.base,
				test.expected)
		}
	}

}

// TestMultipleNamespaces before the same patch
// on two objects have the same name but in a different namespaces
func TestMultipleNamespaces(t *testing.T) {
	tests := []struct {
		name          string
		base          []string
		patch         []string
		expected      []string
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-ns1-ns2",
			base: []string{
				addNamespace("ns1", baseResource(Deployment)),
				addNamespace("ns2", baseResource(Deployment)),
			},
			patch: []string{
				addNamespace("ns1", addLabelAndEnvPatch(Deployment)),
				addNamespace("ns2", addLabelAndEnvPatch(Deployment)),
			},
			errorExpected: false,
			expected: []string{
				addNamespace("ns1", expectedResultSMP()),
				addNamespace("ns2", expectedResultSMP()),
			},
		},
		{
			name: "noschema-ns1-ns2",
			base: []string{
				addNamespace("ns1", baseResource(MyCRD)),
				addNamespace("ns2", baseResource(MyCRD)),
			},
			patch: []string{
				addNamespace("ns1", addLabelAndEnvPatch(MyCRD)),
				addNamespace("ns2", addLabelAndEnvPatch(MyCRD)),
			},
			errorExpected: false,
			expected: []string{
				addNamespace("ns1", expectedResultJMP("")),
				addNamespace("ns2", expectedResultJMP("")),
			},
		},
		{
			name:          "withschema-ns1-ns2",
			base:          []string{addNamespace("ns1", baseResource(Deployment))},
			patch:         []string{addNamespace("ns2", changeImagePatch(Deployment, "nginx:1.7.9"))},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name:          "withschema-nil-ns2",
			base:          []string{baseResource(Deployment)},
			patch:         []string{addNamespace("ns2", changeImagePatch(Deployment, "nginx:1.7.9"))},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name:          "withschema-ns1-nil",
			base:          []string{addNamespace("ns1", baseResource(Deployment))},
			patch:         []string{changeImagePatch(Deployment, "nginx:1.7.9")},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name:          "noschema-ns1-ns2",
			base:          []string{addNamespace("ns1", baseResource(MyCRD))},
			patch:         []string{addNamespace("ns2", changeImagePatch(MyCRD, "nginx:1.7.9"))},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name:          "noschema-nil-ns2",
			base:          []string{baseResource(MyCRD)},
			patch:         []string{addNamespace("ns2", changeImagePatch(MyCRD, "nginx:1.7.9"))},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name:          "noschema-ns1-nil",
			base:          []string{addNamespace("ns1", baseResource(MyCRD))},
			patch:         []string{changeImagePatch(MyCRD, "nginx:1.7.9")},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
	}

	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	for _, test := range tests {
		th.ResetLoaderRoot(fmt.Sprintf("/%s", test.name))
		for idx, patch := range test.patch {
			th.WriteF(fmt.Sprintf("/%s/patch%d.yaml", test.name, idx), patch)
		}

		if test.errorExpected {
			th.RunTransformerAndCheckError(toConfig(test.patch...), strings.Join(test.base, "\n---\n"),
				func(t *testing.T, err error) {
					compareExpectedError(t, test.name, err, test.errorMsg)
				})
		} else {
			th.RunTransformerAndCheckResult(toConfig(test.patch...), strings.Join(test.base, "\n---\n"),
				strings.Join(test.expected, "---\n"))
		}
	}
}

func TestPatchStrategicMergeTransformerPatchDelete1(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  replica: 2
  template:
    $patch: delete
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
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
  replica: 2
`)
}

func TestPatchStrategicMergeTransformerPatchDelete2(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
spec:
  $patch: delete
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
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
`)
}

func TestPatchStrategicMergeTransformerPatchDelete3(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchStrategicMergeTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
metadata:
  name: myDeploy
kind: Deployment
$patch: delete
`)

	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchStrategicMergeTransformer
metadata:
  name: notImportantHere
paths:
- patch.yaml
`, target)

	th.AssertActualEqualsExpected(rm, ``)
}
