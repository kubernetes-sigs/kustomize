// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resid"
	. "sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
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

const genArgOptions = "{nsfx:false,beh:unspecified}"

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
			s:  configMapAsString + genArgOptions,
		},
		{
			in: testDeployment,
			s:  deploymentAsString + genArgOptions,
		},
	}
	for _, test := range tests {
		if test.in.String() != test.s {
			t.Fatalf("Expected %s == %s", test.in.String(), test.s)
		}
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
				resid.Gvk{Version: "v1", Kind: "ConfigMap"}, "winnie", "hundred-acre-wood"),
		},
		{
			in: testDeployment,
			id: resid.NewResId(resid.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, "pooh"),
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
