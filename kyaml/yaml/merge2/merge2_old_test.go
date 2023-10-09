// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	. "sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

const dest = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h1
    i: j
    m: n2
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['c', 'a', 'b']
        env:
        - name: DEMO_GREETING
          value: "Hello from the environment"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
`

func TestMerge_map(t *testing.T) {
	dest := yaml.MustParse(dest)
	src := yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h2
    k: l
    m: n1
`)

	result, err := Merge(src, dest, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListAppend,
	})
	if !assert.NoError(t, err) {
		return
	}
	actual, err := result.String()
	if !assert.NoError(t, err) {
		return
	}

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h2
    i: j
    k: l
    m: n1
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['c', 'a', 'b']
        env:
        - name: DEMO_GREETING
          value: "Hello from the environment"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
`
	b, err := filters.FormatInput(bytes.NewBufferString(expected))
	if !assert.NoError(t, err) {
		return
	}
	expected = b.String()

	b, err = filters.FormatInput(bytes.NewBufferString(actual))
	if !assert.NoError(t, err) {
		return
	}
	actual = b.String()

	assert.Equal(t, expected, actual)
}

func TestMerge_null(t *testing.T) {
	dest := yaml.MustParse(`
kind: Deployment
metadata:
  annotations: null
`)
	src := yaml.MustParse(`
kind: Deployment
`)

	expected, err := filters.FormatInput(bytes.NewBufferString(`
kind: Deployment
metadata:
  annotations: null
`))
	if !assert.NoError(t, err) {
		return
	}

	// Merge the same src several times to test idempotency
	// https://github.com/kubernetes-sigs/kustomize/issues/5031
	for i := 0; i < 3; i++ {
		result, err := Merge(src, dest, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})

		if !assert.NoError(t, err) {
			return
		}
		got, err := result.String()
		if !assert.NoError(t, err) {
			return
		}

		formatted, err := filters.FormatInput(bytes.NewBufferString(got))
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, expected, formatted)
		dest = result
	}
}

func TestMerge_clear(t *testing.T) {
	dest := yaml.MustParse(dest)
	src := yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations: null
`)

	result, err := Merge(src, dest, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListAppend,
	})
	if !assert.NoError(t, err) {
		return
	}
	actual, err := result.String()
	if !assert.NoError(t, err) {
		return
	}

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['c', 'a', 'b']
        env:
        - name: DEMO_GREETING
          value: "Hello from the environment"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
`
	b, err := filters.FormatInput(bytes.NewBufferString(expected))
	if !assert.NoError(t, err) {
		return
	}
	expected = b.String()

	b, err = filters.FormatInput(bytes.NewBufferString(actual))
	if !assert.NoError(t, err) {
		return
	}
	actual = b.String()

	assert.Equal(t, expected, actual)
}

func TestMerge_mapInverse(t *testing.T) {
	dest := yaml.MustParse(dest)
	src := yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h2
    k: l
    m: n1
`)

	result, err := Merge(dest, src, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListAppend,
	})
	if !assert.NoError(t, err) {
		return
	}
	actual, err := result.String()
	if !assert.NoError(t, err) {
		return
	}

	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h1
    i: j
    k: l
    m: n2
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['c', 'a', 'b']
        env:
        - name: DEMO_GREETING
          value: "Hello from the environment"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
`
	b, err := filters.FormatInput(bytes.NewBufferString(expected))
	if !assert.NoError(t, err) {
		return
	}
	expected = b.String()

	b, err = filters.FormatInput(bytes.NewBufferString(actual))
	if !assert.NoError(t, err) {
		return
	}
	actual = b.String()

	assert.Equal(t, expected, actual)
}

func TestMerge_listElem(t *testing.T) {
	dest := yaml.MustParse(dest)
	src := yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        env:
        - name: DEMO_GREETING
          value: "New Demo Greeting"
        - name: NEW_DEMO_VALUE
          value: "Another Env Not In The Dest"
`)

	result, err := Merge(src, dest, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListAppend,
	})
	if !assert.NoError(t, err) {
		return
	}
	actual, err := result.String()
	if !assert.NoError(t, err) {
		return
	}

	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h1
    i: j
    m: n2
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['c', 'a', 'b']
        env:
        - name: DEMO_GREETING
          value: "New Demo Greeting"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
        - name: NEW_DEMO_VALUE
          value: "Another Env Not In The Dest"
`

	b, err := filters.FormatInput(bytes.NewBufferString(expected))
	if !assert.NoError(t, err) {
		return
	}
	expected = b.String()

	b, err = filters.FormatInput(bytes.NewBufferString(actual))
	if !assert.NoError(t, err) {
		return
	}
	actual = b.String()

	assert.Equal(t, expected, actual)
}

func TestMerge_list(t *testing.T) {
	dest := yaml.MustParse(dest)
	src := yaml.MustParse(`
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: nginx
        args: ['e', 'd', 'f']
`)

	result, err := Merge(src, dest, yaml.MergeOptions{
		ListIncreaseDirection: yaml.MergeOptionsListAppend,
	})
	if !assert.NoError(t, err) {
		return
	}
	actual, err := result.String()
	if !assert.NoError(t, err) {
		return
	}

	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
    g: h1
    i: j
    m: n2
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        args: ['e', 'd', 'f']
        env:
        - name: DEMO_GREETING
          value: "Hello from the environment"
        - name: DEMO_FAREWELL
          value: "Such a sweet sorrow"
`

	b, err := filters.FormatInput(bytes.NewBufferString(expected))
	if !assert.NoError(t, err) {
		return
	}
	expected = b.String()

	b, err = filters.FormatInput(bytes.NewBufferString(actual))
	if !assert.NoError(t, err) {
		return
	}
	actual = b.String()

	assert.Equal(t, expected, actual)
}

func TestMerge_commentsKept(t *testing.T) {
	actual, err := MergeStrings(`
a:
  b:
    c: e
`,
		`
a:
  b:
    # header comment
    c: d
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    # header comment
    c: e
`, actual)

	actual, err = MergeStrings(`
a:
  b:
    c: e
`,
		`
a:
  b:
    c: d
    # footer comment
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    c: e
    # footer comment
`, actual)

	actual, err = MergeStrings(`
a:
  b:
    c: e
`,
		`
a:
  b:
    c: d # line comment
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    c: e
`, actual)
}

func TestMerge_commentsOverride(t *testing.T) {
	actual, err := MergeStrings(`
a:
  b:
    # header comment
    c: e
`,
		`
a:
  b:
    # replace comment
    c: d
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    # replace comment
    c: e
`, actual)

	actual, err = MergeStrings(`
a:
  b:
    c: e
    # footer comment
`,
		`
a:
  b:
    c: d
    # replace comment
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    c: e
    # replace comment
`, actual)

	actual, err = MergeStrings(`
a:
  b:
    c: e # line comment
`,
		`
a:
  b:
    c: d # replace comment
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    c: e # line comment
`, actual)

	actual, err = MergeStrings(`
a:
  b:
    c: d # line comment
`,
		`
a:
  b:
    c: d # replace comment
`, true, yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListAppend,
		})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, `a:
  b:
    c: d # line comment
`, actual)
}
