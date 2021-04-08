// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kioutil_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortNodes_moreThan10(t *testing.T) {
	input := `
a: b
---
c: d
---
e: f
---
g: h
---
i: j
---
k: l
---
m: n
---
o: p
---
q: r
---
s: t
---
u: v
---
w: x
---
y: z
`
	actual := &bytes.Buffer{}
	rw := kio.ByteReadWriter{Reader: bytes.NewBufferString(input), Writer: actual}
	nodes, err := rw.Read()
	if !assert.NoError(t, err) {
		t.Fail()
	}

	// randomize the list
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

	// sort them back into their original order
	if !assert.NoError(t, kioutil.SortNodes(nodes)) {
		t.Fail()
	}

	// check the sorted values
	expected := strings.Split(input, "---")
	for i := range nodes {
		a := strings.TrimSpace(nodes[i].MustString())
		b := strings.TrimSpace(expected[i])
		if !assert.Contains(t, a, b) {
			t.Fail()
		}
	}

	if !assert.NoError(t, rw.Write(nodes)) {
		t.Fail()
	}

	assert.Equal(t, strings.TrimSpace(input), strings.TrimSpace(actual.String()))
}

func TestDefaultPathAnnotation(t *testing.T) {
	var tests = []struct {
		dir      string
		input    string // input
		expected string // expected result
		name     string
	}{
		{
			`foo`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'foo/b/bar_a.yaml'
`, `with namespace`},
		{
			`foo`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'foo/bar_a.yaml'
`, `without namespace`},

		{
			``,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'b/bar_a.yaml'
`, `without dir`},
		{
			``,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
`, `skip`},
	}

	for _, s := range tests {
		n := yaml.MustParse(s.input)
		err := kioutil.DefaultPathAnnotation(s.dir, []*yaml.RNode{n})
		if !assert.NoError(t, err, s.name) {
			t.FailNow()
		}
		if !assert.Equal(t, s.expected, n.MustString(), s.name) {
			t.FailNow()
		}
	}
}

func TestDefaultPathAndIndexAnnotation(t *testing.T) {
	var tests = []struct {
		dir      string
		input    string // input
		expected string // expected result
		name     string
	}{
		{
			`foo`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'foo/b/bar_a.yaml'
    config.kubernetes.io/index: '0'
`, `with namespace`},
		{
			`foo`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'foo/bar_a.yaml'
    config.kubernetes.io/index: '0'
`, `without namespace`},

		{
			``,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'b/bar_a.yaml'
    config.kubernetes.io/index: '0'
`, `without dir`},
		{
			``,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/index: '5'
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/index: '5'
`, `skip`},
	}

	for _, s := range tests {
		out := &bytes.Buffer{}
		r := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(s.input),
			Writer:                out,
			KeepReaderAnnotations: true,
			OmitReaderAnnotations: true,
		}
		n, err := r.Read()
		if !assert.NoError(t, err, s.name) {
			t.FailNow()
		}
		if !assert.NoError(t, kioutil.DefaultPathAndIndexAnnotation(s.dir, n), s.name) {
			t.FailNow()
		}
		if !assert.NoError(t, r.Write(n), s.name) {
			t.FailNow()
		}
		if !assert.Equal(t, s.expected, out.String(), s.name) {
			t.FailNow()
		}
	}
}

func TestCreatePathAnnotationValue(t *testing.T) {
	var tests = []struct {
		dir      string
		meta     yaml.ResourceMeta // input
		expected string            // expected result
		name     string
	}{
		{
			`dir`,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "foo",
				},
				ObjectMeta: yaml.ObjectMeta{
					NameMeta: yaml.NameMeta{
						Name: "bar", Namespace: "baz",
					},
				},
			},
			`dir/baz/foo_bar.yaml`, `with namespace`,
		},
		{
			``,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "foo",
				},
				ObjectMeta: yaml.ObjectMeta{
					NameMeta: yaml.NameMeta{
						Name: "bar", Namespace: "baz",
					},
				},
			},
			`baz/foo_bar.yaml`, `without dir`,
		},
		{
			`dir`,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "foo",
				},
				ObjectMeta: yaml.ObjectMeta{
					NameMeta: yaml.NameMeta{Name: "bar"},
				},
			},
			`dir/foo_bar.yaml`, `without namespace`,
		},
		{
			``,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "foo",
				},
				ObjectMeta: yaml.ObjectMeta{
					NameMeta: yaml.NameMeta{Name: "bar"},
				},
			},
			`foo_bar.yaml`, `without namespace or dir`,
		},
		{
			``,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "foo",
				},
				ObjectMeta: yaml.ObjectMeta{},
			},
			`foo_.yaml`, `without namespace, dir or name`,
		},
		{
			``,
			yaml.ResourceMeta{
				TypeMeta: yaml.TypeMeta{
					APIVersion: "apps/v1",
				},
				ObjectMeta: yaml.ObjectMeta{},
			},
			`_.yaml`, `without any`,
		},
	}

	for _, s := range tests {
		p := kioutil.CreatePathAnnotationValue(s.dir, s.meta)
		if !assert.Equal(t, s.expected, p, s.name) {
			t.FailNow()
		}
	}
}

func TestDuplicatePathAndIndexAnnotation(t *testing.T) {
	tests := map[string]struct {
		input       string // input
		expectedErr string // expected result
	}{
		"duplicate": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'my/path/custom.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/custom.yaml'
    config.kubernetes.io/index: '0'
`,
			expectedErr: "duplicate path and index",
		},
		"duplicate path, not index": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'my/path/custom.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/custom.yaml'
    config.kubernetes.io/index: '1'
`,
		},
		"duplicate index, not path": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'my/path/a.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '0'
`,
		},
		"larger number of resources with duplicate": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'my/path/a.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/a.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '2'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/c.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/c.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '1'
`,
			expectedErr: "duplicate path and index",
		},
		"larger number of resources without duplicates": {
			input: `apiVersion: v1
kind: Custom
metadata:
  name: a
  annotations:
    config.kubernetes.io/path: 'my/path/a.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/a.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '2'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/c.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/c.yaml'
    config.kubernetes.io/index: '1'
---
apiVersion: v1
kind: Custom
metadata:
  name: b
  annotations:
    config.kubernetes.io/path: 'my/path/b.yaml'
    config.kubernetes.io/index: '3'
`,
		},
	}
	for _, tc := range tests {
		out := &bytes.Buffer{}
		r := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(tc.input),
			Writer:                out,
			KeepReaderAnnotations: true,
			OmitReaderAnnotations: true,
		}
		n, err := r.Read()
		if err != nil {
			t.FailNow()
		}
		err = kioutil.ErrorIfDuplicateAnnotation(n)
		if err != nil && tc.expectedErr == "" {
			t.Errorf("unexpected error %s", err.Error())
			t.FailNow()
		}
		if tc.expectedErr != "" && err == nil {
			t.Errorf("expected error %s", tc.expectedErr)
			t.FailNow()
		}
		if tc.expectedErr != "" && !strings.Contains(err.Error(), tc.expectedErr) {
			t.FailNow()
		}
	}
}
