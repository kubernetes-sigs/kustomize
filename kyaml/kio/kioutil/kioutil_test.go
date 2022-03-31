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
    internal.config.kubernetes.io/path: 'foo/b/bar_a.yaml'
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
    internal.config.kubernetes.io/path: 'foo/bar_a.yaml'
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
    internal.config.kubernetes.io/path: 'b/bar_a.yaml'
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
    internal.config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/path: 'a/b.yaml'
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
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
    internal.config.kubernetes.io/path: 'foo/b/bar_a.yaml'
    config.kubernetes.io/path: 'foo/b/bar_a.yaml'
    internal.config.kubernetes.io/index: '0'
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
    internal.config.kubernetes.io/path: 'foo/bar_a.yaml'
    config.kubernetes.io/path: 'foo/bar_a.yaml'
    internal.config.kubernetes.io/index: '0'
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
    internal.config.kubernetes.io/path: 'b/bar_a.yaml'
    config.kubernetes.io/path: 'b/bar_a.yaml'
    internal.config.kubernetes.io/index: '0'
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
    internal.config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    config.kubernetes.io/index: '5'
`,
			`apiVersion: v1
kind: Bar
metadata:
  name: a
  namespace: b
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
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

func TestCopyLegacyAnnotations(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/index: '5'
`,
			expected: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
`,
			expected: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/index: '5'
`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
 name: foobar
 annotations:
   internal.config.kubernetes.io/path: 'a/b.yaml'
   config.kubernetes.io/path: 'c/d.yaml'
`,
			expected: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    config.kubernetes.io/path: 'c/d.yaml'
`,
		},
	}

	for _, tc := range tests {
		rw := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(tc.input),
			OmitReaderAnnotations: true,
		}
		nodes, err := rw.Read()
		assert.NoError(t, err)
		assert.NoError(t, kioutil.CopyLegacyAnnotations(nodes[0]))
		assert.Equal(t, tc.expected, nodes[0].MustString())
	}
}

func TestCopyInternalAnnotations(t *testing.T) {
	var tests = []struct {
		input      string
		exclusions []kioutil.AnnotationKey
		expected   string
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: src
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar'
---
apiVersion: v1
kind: Foo
metadata:
  name: dst
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
`,
			expected: `apiVersion: v1
kind: Foo
metadata:
  name: dst
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar'
`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: src
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar-src'
---
apiVersion: v1
kind: Foo
metadata:
  name: dst
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
    internal.config.kubernetes.io/foo: 'bar-dst'
`,
			exclusions: []kioutil.AnnotationKey{
				kioutil.PathAnnotation,
				kioutil.IndexAnnotation,
			},
			expected: `apiVersion: v1
kind: Foo
metadata:
  name: dst
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
    internal.config.kubernetes.io/foo: 'bar-src'
`,
		},
	}

	for _, tc := range tests {
		rw := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(tc.input),
			OmitReaderAnnotations: true,
		}
		nodes, err := rw.Read()
		assert.NoError(t, err)
		assert.NoError(t, kioutil.CopyInternalAnnotations(nodes[0], nodes[1], tc.exclusions...))
		assert.Equal(t, tc.expected, nodes[1].MustString())
	}
}

func TestConfirmInternalAnnotationUnchanged(t *testing.T) {
	var tests = []struct {
		input       string
		exclusions  []kioutil.AnnotationKey
		expectedErr string
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo-1
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
---
apiVersion: v1
kind: Foo
metadata:
  name: foo-2
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
`,
			expectedErr: `internal annotations differ: internal.config.kubernetes.io/index, internal.config.kubernetes.io/path`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo-1
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
---
apiVersion: v1
kind: Foo
metadata:
  name: foo-2
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
`,
			exclusions: []kioutil.AnnotationKey{
				kioutil.PathAnnotation,
				kioutil.IndexAnnotation,
			},
			expectedErr: ``,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo-1
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar-1'
---
apiVersion: v1
kind: Foo
metadata:
  name: foo-2
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
    internal.config.kubernetes.io/foo: 'bar-2'
`,
			exclusions: []kioutil.AnnotationKey{
				kioutil.PathAnnotation,
				kioutil.IndexAnnotation,
			},
			expectedErr: `internal annotations differ: internal.config.kubernetes.io/foo`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo-1
  annotations:
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar-1'
---
apiVersion: v1
kind: Foo
metadata:
  name: foo-2
  annotations:
    internal.config.kubernetes.io/path: 'c/d.yaml'
    internal.config.kubernetes.io/index: '10'
    internal.config.kubernetes.io/foo: 'bar-1'
`,
			expectedErr: `internal annotations differ: internal.config.kubernetes.io/index, internal.config.kubernetes.io/path`,
		},
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foo-1
  annotations:
    internal.config.kubernetes.io/a: 'b'
    internal.config.kubernetes.io/c: 'd'
---
apiVersion: v1
kind: Foo
metadata:
  name: foo-2
  annotations:
    internal.config.kubernetes.io/e: 'f'
    internal.config.kubernetes.io/g: 'h'
`,
			expectedErr: `internal annotations differ: ` +
				`internal.config.kubernetes.io/a, internal.config.kubernetes.io/c, ` +
				`internal.config.kubernetes.io/e, internal.config.kubernetes.io/g`,
		},
	}

	for _, tc := range tests {
		rw := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(tc.input),
			OmitReaderAnnotations: true,
		}
		nodes, err := rw.Read()
		assert.NoError(t, err)
		err = kioutil.ConfirmInternalAnnotationUnchanged(nodes[0], nodes[1], tc.exclusions...)
		if tc.expectedErr == "" {
			assert.NoError(t, err)
		} else {
			if err == nil {
				t.Fatalf("expected error: %s\n", tc.expectedErr)
			}
			assert.Equal(t, tc.expectedErr, err.Error())
		}
	}
}

func TestGetInternalAnnotations(t *testing.T) {
	var tests = []struct {
		input    string
		expected map[string]string
	}{
		{
			input: `apiVersion: v1
kind: Foo
metadata:
  name: foobar
  annotations:
    foo: bar
    internal.config.kubernetes.io/path: 'a/b.yaml'
    internal.config.kubernetes.io/index: '5'
    internal.config.kubernetes.io/foo: 'bar'
`,
			expected: map[string]string{
				"internal.config.kubernetes.io/path":  "a/b.yaml",
				"internal.config.kubernetes.io/index": "5",
				"internal.config.kubernetes.io/foo":   "bar",
			},
		},
	}

	for _, tc := range tests {
		rw := kio.ByteReadWriter{
			Reader:                bytes.NewBufferString(tc.input),
			OmitReaderAnnotations: true,
		}
		nodes, err := rw.Read()
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, kioutil.GetInternalAnnotations(nodes[0]))
	}
}
