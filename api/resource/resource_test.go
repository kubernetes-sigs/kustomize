// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/provider"
	. "sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var factory = provider.NewDefaultDepProvider().GetResourceFactory()

var testConfigMap = factory.FromMap(
	map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "winnie",
			"namespace": "hundred-acre-wood",
		},
	})

//nolint:gosec
const configMapAsString = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"winnie","namespace":"hundred-acre-wood"}}`

var testDeployment = factory.FromMap(
	map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "pooh",
		},
	})

const deploymentAsString = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"pooh"}}`

func TestAsYAML(t *testing.T) {
	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: pooh
`
	yaml, err := testDeployment.AsYAML()
	if err != nil {
		t.Fatal(err)
	}
	if string(yaml) != expected {
		t.Fatalf("--- expected\n%s\n--- got\n%s\n", expected, string(yaml))
	}
}

func TestResourceString(t *testing.T) {
	tests := []struct {
		in *Resource
		s  string
	}{
		{
			in: testConfigMap,
			s:  configMapAsString,
		},
		{
			in: testDeployment,
			s:  deploymentAsString,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.in.String(), test.s)
	}
}

func TestResourceId(t *testing.T) {
	tests := []struct {
		in *Resource
		id resid.ResId
	}{
		{
			in: testConfigMap,
			id: resid.NewResIdWithNamespace(
				resid.NewGvk("", "v1", "ConfigMap"),
				"winnie", "hundred-acre-wood"),
		},
		{
			in: testDeployment,
			id: resid.NewResId(
				resid.NewGvk("apps", "v1", "Deployment"), "pooh"),
		},
	}
	for _, test := range tests {
		if test.in.OrgId() != test.id {
			t.Fatalf("Expected %v, but got %v\n", test.id, test.in.OrgId())
		}
	}
}

func TestDeepCopy(t *testing.T) {
	r := factory.FromMap(
		map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "pooh",
			},
		})
	r.AppendRefBy(resid.NewResId(resid.Gvk{Group: "somegroup", Kind: "MyKind"}, "random"))

	var1 := types.Var{
		Name: "SERVICE_ONE",
		ObjRef: types.Target{
			Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
			Name: "backendOne"},
	}
	r.AppendRefVarName(var1)

	cr := r.DeepCopy()
	if !reflect.DeepEqual(r, cr) {
		t.Errorf("expected %v\nbut got%v", r, cr)
	}
}

func TestApplySmPatch_1(t *testing.T) {
	resource, err := factory.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: bingo
spec:
  replicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        image: nginx
        name: nginx
        ports:
        - containerPort: 80
`))
	assert.NoError(t, err)
	patch, err := factory.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: baseprefix-mungebot
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        ports:
        - containerPort: 777
`))
	assert.NoError(t, err)

	assert.NoError(t, resource.ApplySmPatch(patch))
	bytes, err := resource.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    baseAnno: This is a base annotation
  labels:
    app: mungebot
    foo: bar
  name: bingo
spec:
  replicas: 1
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - env:
        - name: foo
          value: bar
        image: nginx
        name: nginx
        ports:
        - containerPort: 777
        - containerPort: 80
`, string(bytes))
}

func TestApplySmPatch_2(t *testing.T) {
	resource, err := factory.FromBytes([]byte(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    A: X
    B: Y
`))
	assert.NoError(t, err)
	patch, err := factory.FromBytes([]byte(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
    D: W
  baz:
    hello: world
`))
	assert.NoError(t, err)
	assert.NoError(t, resource.ApplySmPatch(patch))
	bytes, err := resource.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: example.com/v1
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
`, string(bytes))
}

func TestApplySmPatch_3(t *testing.T) {
	resource, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)
	patch, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`))
	assert.NoError(t, err)
	assert.NoError(t, resource.ApplySmPatch(patch))
	bytes, err := resource.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 999
`, string(bytes))
}

func TestApplySmPatchShouldOutputListItemsInCorrectOrder(t *testing.T) {
	cases := []struct {
		name           string
		skip           bool
		patch          string
		expectedOutput string
	}{
		{
			name: "Order should not change when patch has foo only",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: foo
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
  - name: foo
  - name: bar
`,
		},
		{
			name: "Order changes when patch has bar only",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: bar
`,
			// This test records current behavior, but this behavior might be undesirable.
			// If so, feel free to change the test to pass with some improved algorithm.
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
  - name: bar
  - name: foo
`,
		},
		{
			name: "Order should not change and should include a new item at the beginning when patch has a new list item",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: baz
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
  - name: baz
  - name: foo
  - name: bar
`,
		},
		{
			name: "Order should not change when patch has foo and bar in same order",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: foo
    - name: bar
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
  - name: foo
  - name: bar
`,
		},
		{
			name: "Order should change when patch has foo and bar in different order",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: bar
    - name: foo
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
  - name: bar
  - name: foo
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			resource, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  initContainers:
    - name: foo
    - name: bar
`))
			assert.NoError(t, err)

			patch, err := factory.FromBytes([]byte(tc.patch))
			assert.NoError(t, err)
			assert.NoError(t, resource.ApplySmPatch(patch))
			bytes, err := resource.AsYAML()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, string(bytes))
		})
	}
}

func TestApplySmPatchShouldOutputPrimitiveListItemsInCorrectOrder(t *testing.T) {
	cases := []struct {
		name           string
		skip           bool
		patch          string
		expectedOutput string
	}{
		{
			name: "Order should not change when patch has foo only",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
  finalizers: ["foo"]
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  finalizers:
  - foo
  - bar
  name: test
`,
		},
		{
			name: "Order should not change when patch has bar only",
			skip: true, // TODO: This test should pass but fails currently. Fix the problem and unskip this test
			patch: `apiVersion: v1
kind: Pod
metadata:
 name: test
 finalizers: ["bar"]
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
 finalizers:
 - foo
 - bar
 name: test
`,
		},
		{
			name: "Order should not change and should include a new item at the beginning when patch has a new list item",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
  finalizers: ["baz"]
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  finalizers:
  - baz
  - foo
  - bar
  name: test
`,
		},
		{
			name: "Order should not change when patch has foo and bar in same order",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
  finalizers: ["foo", "bar"]
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  finalizers:
  - foo
  - bar
  name: test
`,
		},
		{
			name: "Order should change when patch has foo and bar in different order",
			patch: `apiVersion: v1
kind: Pod
metadata:
  name: test
  finalizers: ["bar", "foo"]
`,
			expectedOutput: `apiVersion: v1
kind: Pod
metadata:
  finalizers:
  - bar
  - foo
  name: test
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			resource, err := factory.FromBytes([]byte(`
kind: Pod
metadata:
  name: test
  finalizers: ["foo", "bar"]
`))
			assert.NoError(t, err)

			patch, err := factory.FromBytes([]byte(tc.patch))
			assert.NoError(t, err)
			assert.NoError(t, resource.ApplySmPatch(patch))
			bytes, err := resource.AsYAML()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOutput, string(bytes))
		})
	}
}

func TestMergeDataMapFrom(t *testing.T) {
	resource, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: BlahBlah
metadata:
  name: clown
data:
  fruit: pear
`))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	patch, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Whatever
metadata:
  name: spaceship
data:
  spaceship: enterprise
`))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	resource.MergeDataMapFrom(patch)
	bytes, err := resource.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
data:
  fruit: pear
  spaceship: enterprise
kind: BlahBlah
metadata:
  name: clown
`, string(bytes))
}

func TestApplySmPatch_SwapOrder(t *testing.T) {
	s1 := `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    B:
    C: Z
`
	s2 := `
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
`
	expected := `apiVersion: example.com/v1
kind: Foo
metadata:
  name: my-foo
spec:
  bar:
    C: Z
    D: W
  baz:
    hello: world
`
	r1, err := factory.FromBytes([]byte(s1))
	assert.NoError(t, err)
	r2, err := factory.FromBytes([]byte(s2))
	assert.NoError(t, err)
	assert.NoError(t, r1.ApplySmPatch(r2))
	bytes, err := r1.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, expected, string(bytes))

	r1, _ = factory.FromBytes([]byte(s1))
	r2, _ = factory.FromBytes([]byte(s2))
	assert.NoError(t, r2.ApplySmPatch(r1))
	bytes, err = r2.AsYAML()
	assert.NoError(t, err)
	assert.Equal(t, expected, string(bytes))
}

func TestApplySmPatch(t *testing.T) {
	const (
		myDeployment = "Deployment"
		myCRD        = "myCRD"
	)

	tests := map[string]struct {
		base          string
		patch         []string
		expected      string
		errorExpected bool
		errorMsg      string
	}{
		"withschema-label-image-container": {
			base: baseResource(myDeployment),
			patch: []string{
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:latest"),
				addContainerAndEnvPatch(myDeployment),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(myDeployment, false),
		},
		"withschema-image-container-label": {
			base: baseResource(myDeployment),
			patch: []string{
				changeImagePatch(myDeployment, "nginx:latest"),
				addContainerAndEnvPatch(myDeployment),
				addLabelAndEnvPatch(myDeployment),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(myDeployment, true),
		},
		"withschema-container-label-image": {
			base: baseResource(myDeployment),
			patch: []string{
				addContainerAndEnvPatch(myDeployment),
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:latest"),
			},
			errorExpected: false,
			expected:      expectedResultMultiPatch(myDeployment, true),
		},
		"noschema-label-image-container": {
			base: baseResource(myCRD),
			patch: []string{
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:latest"),
				addContainerAndEnvPatch(myCRD),
			},
			// Might be better if this complained about patch conflict.
			// See plugin/builtin/patchstrategicmergetransformer/psmt_test.go
			expected: `apiVersion: apps/v1
kind: myCRD
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
        name: nginx
      - image: anotherimage
        name: anothercontainer
`,
		},
		"noschema-image-container-label": {
			base: baseResource(myCRD),
			patch: []string{
				changeImagePatch(myCRD, "nginx:latest"),
				addContainerAndEnvPatch(myCRD),
				addLabelAndEnvPatch(myCRD),
			},
			// Might be better if this complained about patch conflict.
			expected: `apiVersion: apps/v1
kind: myCRD
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
`,
		},
		"noschema-container-label-image": {
			base: baseResource(myCRD),
			patch: []string{
				addContainerAndEnvPatch(myCRD),
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:latest"),
			},
			// Might be better if this complained about patch conflict.
			expected: `apiVersion: apps/v1
kind: myCRD
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
      - image: nginx:latest
        name: nginx
`,
		},

		"withschema-label-latest-someV-01": {
			base: baseResource(myDeployment),
			patch: []string{
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:latest"),
				changeImagePatch(myDeployment, "nginx:1.7.9"),
			},
			expected: `apiVersion: apps/v1
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
        image: nginx:1.7.9
        name: nginx
`,
		},
		"withschema-latest-label-someV-02": {
			base: baseResource(myDeployment),
			patch: []string{
				changeImagePatch(myDeployment, "nginx:latest"),
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:1.7.9"),
			},
			expected: `apiVersion: apps/v1
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
        image: nginx:1.7.9
        name: nginx
`,
		},
		"withschema-latest-label-someV-03": {
			base: baseResource(myDeployment),
			patch: []string{
				changeImagePatch(myDeployment, "nginx:1.7.9"),
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:latest"),
			},
			expected: `apiVersion: apps/v1
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
        image: nginx:latest
        name: nginx
`,
		},
		"withschema-latest-label-someV-04": {
			base: baseResource(myDeployment),
			patch: []string{
				changeImagePatch(myDeployment, "nginx:1.7.9"),
				changeImagePatch(myDeployment, "nginx:latest"),
				addLabelAndEnvPatch(myDeployment),
				changeImagePatch(myDeployment, "nginx:nginx"),
			},
			expected: `apiVersion: apps/v1
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
        image: nginx:nginx
        name: nginx
`,
		},
		"noschema-latest-label-someV-01": {
			base: baseResource(myCRD),
			patch: []string{
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:latest"),
				changeImagePatch(myCRD, "nginx:1.7.9"),
			},
			expected: `apiVersion: apps/v1
kind: myCRD
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
      - image: nginx:1.7.9
        name: nginx
`,
		},
		"noschema-latest-label-someV-02": {
			base: baseResource(myCRD),
			patch: []string{
				changeImagePatch(myCRD, "nginx:latest"),
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:1.7.9"),
			},
			expected: expectedResultJMP("nginx:1.7.9"),
		},
		"noschema-latest-label-someV-03": {
			base: baseResource(myCRD),
			patch: []string{
				changeImagePatch(myCRD, "nginx:1.7.9"),
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:latest"),
			},
			expected: `apiVersion: apps/v1
kind: myCRD
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
      - image: nginx:latest
        name: nginx
`,
		},
		"noschema-latest-label-someV-04": {
			base: baseResource(myCRD),
			patch: []string{
				changeImagePatch(myCRD, "nginx:1.7.9"),
				changeImagePatch(myCRD, "nginx:latest"),
				addLabelAndEnvPatch(myCRD),
				changeImagePatch(myCRD, "nginx:nginx"),
			},
			expected: `apiVersion: apps/v1
kind: myCRD
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
      - image: nginx:nginx
        name: nginx
`,
		},
	}

	for name, test := range tests {
		resource, err := factory.FromBytes([]byte(test.base))
		assert.NoError(t, err)
		for _, p := range test.patch {
			patch, err := factory.FromBytes([]byte(p))
			assert.NoError(t, err, name)
			assert.NoError(t, resource.ApplySmPatch(patch), name)
		}
		bytes, err := resource.AsYAML()
		if test.errorExpected {
			assert.Error(t, err, name)
		} else {
			assert.NoError(t, err, name)
			assert.Equal(t, test.expected, string(bytes), name)
		}
	}
}

func TestResourceStorePreviousId(t *testing.T) {
	tests := map[string]struct {
		input    string
		newName  string
		newNs    string
		expected string
	}{
		"default namespace, first previous name": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  name: oldName
`,
			newName: "newName",
			newNs:   "",
			expected: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret
    internal.config.kubernetes.io/previousNames: oldName
    internal.config.kubernetes.io/previousNamespaces: default
  name: newName
`,
		},

		"default namespace, second previous name": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret
    internal.config.kubernetes.io/previousNames: oldName
    internal.config.kubernetes.io/previousNamespaces: default
  name: oldName2
`,
			newName: "newName",
			newNs:   "",
			expected: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret,Secret
    internal.config.kubernetes.io/previousNames: oldName,oldName2
    internal.config.kubernetes.io/previousNamespaces: default,default
  name: newName
`,
		},

		"non-default namespace": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret
    internal.config.kubernetes.io/previousNames: oldName
    internal.config.kubernetes.io/previousNamespaces: default
  name: oldName2
  namespace: oldNamespace
`,
			newName: "newName",
			newNs:   "newNamespace",
			expected: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret,Secret
    internal.config.kubernetes.io/previousNames: oldName,oldName2
    internal.config.kubernetes.io/previousNamespaces: default,oldNamespace
  name: newName
  namespace: newNamespace
`,
		},
	}
	factory := provider.NewDefaultDepProvider().GetResourceFactory()
	for i := range tests {
		test := tests[i]
		t.Run(i, func(t *testing.T) {
			resources, err := factory.SliceFromBytes([]byte(test.input))
			if !assert.NoError(t, err) || len(resources) == 0 {
				t.FailNow()
			}
			r := resources[0]
			r.StorePreviousId()
			r.SetName(test.newName)
			if test.newNs != "" {
				r.SetNamespace(test.newNs)
			}
			bytes, err := r.AsYAML()
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			assert.Equal(t, test.expected, string(bytes))
		})
	}
}

func TestResource_PrevIds(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []resid.ResId
	}{
		"no previous IDs": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  name: name
`,
			expected: nil,
		},

		"one previous ID": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret
    internal.config.kubernetes.io/previousNames: oldName
    internal.config.kubernetes.io/previousNamespaces: default
  name: newName
`,
			expected: []resid.ResId{
				{
					Gvk:       resid.Gvk{Group: "apps", Version: "v1", Kind: "Secret"},
					Name:      "oldName",
					Namespace: resid.DefaultNamespace,
				},
			},
		},

		"two ids": {
			input: `apiVersion: apps/v1
kind: Secret
metadata:
  annotations:
    internal.config.kubernetes.io/previousKinds: Secret,Secret
    internal.config.kubernetes.io/previousNames: oldName,oldName2
    internal.config.kubernetes.io/previousNamespaces: default,oldNamespace
  name: newName
  namespace: newNamespace
`,
			expected: []resid.ResId{
				{
					Gvk:       resid.Gvk{Group: "apps", Version: "v1", Kind: "Secret"},
					Name:      "oldName",
					Namespace: resid.DefaultNamespace,
				},
				{
					Gvk:       resid.Gvk{Group: "apps", Version: "v1", Kind: "Secret"},
					Name:      "oldName2",
					Namespace: "oldNamespace",
				},
			},
		},
	}
	factory := provider.NewDefaultDepProvider().GetResourceFactory()
	for i := range tests {
		test := tests[i]
		t.Run(i, func(t *testing.T) {
			resources, err := factory.SliceFromBytes([]byte(test.input))
			if !assert.NoError(t, err) || len(resources) == 0 {
				t.FailNow()
			}
			r := resources[0]
			assert.Equal(t, test.expected, r.PrevIds())
		})
	}
}

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
	return fmt.Sprintf(`
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
           value: SOMEVALUE`, kind)
}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a second container in the container list
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func addContainerAndEnvPatch(kind string) string {
	return fmt.Sprintf(`
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
        image: anotherimage`, kind)
}

// addContainerAndEnvPatch produces a patch object which replaces
// the value of the image field in the first/nginx container
// Note that for SMP/WithSchema merge, the name:nginx entry
// is mandatory
func changeImagePatch(kind string, newImage string) string {
	return fmt.Sprintf(`
apiVersion: apps/v1
kind: %s
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: %s`, kind, newImage)
}

// utility method to build the expected result of a multipatch
// the order of the patches still have influence especially
// in the insertion location within arrays.
func expectedResultMultiPatch(kind string, reversed bool) string {
	pattern := `apiVersion: apps/v1
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
        %s
        image: nginx:latest
        name: nginx
      - image: anotherimage
        name: anothercontainer
`
	if reversed {
		return fmt.Sprintf(pattern, kind, `- name: SOMEENV
          value: SOMEVALUE
        - name: ANOTHERENV
          value: ANOTHERVALUE`)
	}
	return fmt.Sprintf(pattern, kind, `- name: ANOTHERENV
          value: ANOTHERVALUE
        - name: SOMEENV
          value: SOMEVALUE`)
}

// utility method building the expected output of a JMP.
// imagename parameter allows to build a result consistent
// with the JMP behavior which basically overrides the
// entire "containers" list.
func expectedResultJMP(imagename string) string {
	if imagename == "" {
		return `apiVersion: apps/v1
kind: myCRD
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
	}
	return fmt.Sprintf(`apiVersion: apps/v1
kind: myCRD
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
`, imagename)
}

func TestSameEndingSubarray(t *testing.T) {
	testCases := map[string]struct {
		a        []string
		b        []string
		expected bool
	}{
		"both nil": {
			expected: true,
		},
		"one nil": {
			b:        []string{},
			expected: true,
		},
		"both empty": {
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		"no1": {
			a:        []string{"a"},
			b:        []string{},
			expected: false,
		},
		"no2": {
			a:        []string{"b", "a"},
			b:        []string{"b"},
			expected: false,
		},
		"yes1": {
			a:        []string{"a", "b"},
			b:        []string{"b"},
			expected: true,
		},
		"yes2": {
			a:        []string{"a", "b", "c"},
			b:        []string{"b", "c"},
			expected: true,
		},
		"yes3": {
			a:        []string{"a", "b", "c", "d", "e", "f"},
			b:        []string{"f"},
			expected: true,
		},
	}
	for n := range testCases {
		tc := testCases[n]
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, tc.expected, utils.SameEndingSubSlice(tc.a, tc.b))
		})
	}
}

func TestGetGvk(t *testing.T) {
	r, err := factory.FromBytes([]byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)

	gvk := r.GetGvk()
	expected := "apps"
	actual := gvk.Group
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "v1"
	actual = gvk.Version
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "Deployment"
	actual = gvk.Kind
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
func TestSetGvk(t *testing.T) {
	r, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)
	r.SetGvk(resid.GvkFromString("knd.ver.grp"))
	gvk := r.GetGvk()
	if expected, actual := "grp", gvk.Group; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "ver", gvk.Version; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "knd", gvk.Kind; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestRefBy(t *testing.T) {
	r, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)
	r.AppendRefBy(resid.FromString("knd1.ver1.gr1/name1.ns1"))
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
  annotations:
    internal.config.kubernetes.io/refBy: 'knd1.ver1.gr1/name1.ns1'
spec:
  numReplicas: 1
`, r.RNode.MustString())
	assert.Equal(t, r.GetRefBy(), []resid.ResId{resid.FromString("knd1.ver1.gr1/name1.ns1")})

	r.AppendRefBy(resid.FromString("knd2.ver2.gr2/name2.ns2"))
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
  annotations:
    internal.config.kubernetes.io/refBy: 'knd1.ver1.gr1/name1.ns1,knd2.ver2.gr2/name2.ns2'
spec:
  numReplicas: 1
`, r.RNode.MustString())
	assert.Equal(t, []resid.ResId{
		resid.FromString("knd1.ver1.gr1/name1.ns1"),
		resid.FromString("knd2.ver2.gr2/name2.ns2"),
	}, r.GetRefBy())
}

func TestOrigin(t *testing.T) {
	r, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)
	origin := &Origin{
		Path: "deployment.yaml",
		Repo: "github.com/myrepo",
		Ref:  "master",
	}
	assert.NoError(t, r.SetOrigin(origin))
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
  annotations:
    config.kubernetes.io/origin: |
      path: deployment.yaml
      repo: github.com/myrepo
      ref: master
spec:
  numReplicas: 1
`, r.MustString())
	or, err := r.GetOrigin()
	assert.NoError(t, err)
	assert.Equal(t, origin, or)
}

func TestTransformations(t *testing.T) {
	r, err := factory.FromBytes([]byte(`
apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`))
	assert.NoError(t, err)
	origin1 := &Origin{
		Repo:         "github.com/myrepo",
		Ref:          "master",
		ConfiguredIn: "config.yaml",
		ConfiguredBy: kyaml.ResourceIdentifier{
			TypeMeta: kyaml.TypeMeta{
				APIVersion: "builtin",
				Kind:       "Generator",
			},
			NameMeta: kyaml.NameMeta{
				Name:      "my-name",
				Namespace: "my-namespace",
			},
		},
	}
	origin2 := &Origin{
		ConfiguredIn: "../base/config.yaml",
		ConfiguredBy: kyaml.ResourceIdentifier{
			TypeMeta: kyaml.TypeMeta{
				APIVersion: "builtin",
				Kind:       "Generator",
			},
			NameMeta: kyaml.NameMeta{
				Name:      "my-name",
				Namespace: "my-namespace",
			},
		},
	}
	assert.NoError(t, r.AddTransformation(origin1))
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - repo: github.com/myrepo
        ref: master
        configuredIn: config.yaml
        configuredBy:
          apiVersion: builtin
          kind: Generator
          name: my-name
          namespace: my-namespace
spec:
  numReplicas: 1
`, r.MustString())
	assert.NoError(t, r.AddTransformation(origin2))
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - repo: github.com/myrepo
        ref: master
        configuredIn: config.yaml
        configuredBy:
          apiVersion: builtin
          kind: Generator
          name: my-name
          namespace: my-namespace
      - configuredIn: ../base/config.yaml
        configuredBy:
          apiVersion: builtin
          kind: Generator
          name: my-name
          namespace: my-namespace
spec:
  numReplicas: 1
`, r.MustString())
	transformations, err := r.GetTransformations()
	assert.NoError(t, err)
	assert.Equal(t, Transformations{origin1, origin2}, transformations)
	assert.NoError(t, r.ClearTransformations())
	assert.Equal(t, `apiVersion: v1
kind: Deployment
metadata:
  name: clown
spec:
  numReplicas: 1
`, r.MustString())
}
