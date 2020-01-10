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
			yaml.ResourceMeta{Kind: "foo",
				APIVersion: "apps/v1",
				ObjectMeta: yaml.ObjectMeta{Name: "bar", Namespace: "baz"},
			},
			`dir/baz/foo_bar.yaml`, `with namespace`,
		},
		{
			``,
			yaml.ResourceMeta{Kind: "foo",
				APIVersion: "apps/v1",
				ObjectMeta: yaml.ObjectMeta{Name: "bar", Namespace: "baz"},
			},
			`baz/foo_bar.yaml`, `without dir`,
		},
		{
			`dir`,
			yaml.ResourceMeta{Kind: "foo",
				APIVersion: "apps/v1",
				ObjectMeta: yaml.ObjectMeta{Name: "bar"},
			},
			`dir/foo_bar.yaml`, `without namespace`,
		},
		{
			``,
			yaml.ResourceMeta{Kind: "foo",
				APIVersion: "apps/v1",
				ObjectMeta: yaml.ObjectMeta{Name: "bar"},
			},
			`foo_bar.yaml`, `without namespace or dir`,
		},
		{
			``,
			yaml.ResourceMeta{Kind: "foo",
				APIVersion: "apps/v1",
				ObjectMeta: yaml.ObjectMeta{},
			},
			`foo_.yaml`, `without namespace, dir or name`,
		},
		{
			``,
			yaml.ResourceMeta{
				APIVersion: "apps/v1",
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
