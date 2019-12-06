// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	. "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestResourceNode_SetValue(t *testing.T) {
	instance := *NewScalarRNode("foo")
	copy := instance
	instance.SetYNode(&yaml.Node{Kind: yaml.ScalarNode, Value: "bar"})
	assert.Equal(t, `bar
`, assertNoErrorString(t)(copy.String()))
	assert.Equal(t, `bar
`, assertNoErrorString(t)(instance.String()))

	instance = *NewScalarRNode("foo")
	copy = instance
	instance.SetYNode(nil)
	instance.SetYNode(&yaml.Node{Kind: yaml.ScalarNode, Value: "bar"})
	assert.Equal(t, `foo
`, assertNoErrorString(t)(copy.String()))
	assert.Equal(t, `bar
`, assertNoErrorString(t)(instance.String()))
}

func TestAppend(t *testing.T) {
	s := `n: o
a: b
c: d
`
	node, err := Parse(s)
	assert.NoError(t, err)
	rn, err := node.Pipe(Append(NewScalarRNode("").YNode()))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}
	assert.Nil(t, rn)

	s = `- a
- b
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(Append())
	assert.NoError(t, err)
	assert.Nil(t, rn)
}

func TestClearField_Fn(t *testing.T) {
	s := `n: o
a: b
c: d
`
	node, err := Parse(s)
	assert.NoError(t, err)
	rn, err := node.Pipe(FieldClearer{Name: "a"})
	assert.NoError(t, err)
	assert.Equal(t, "n: o\nc: d\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "b\n", assertNoErrorString(t)(rn.String()))

	s = `n: o
a: b
c: d
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "n"})
	assert.NoError(t, err)
	assert.Equal(t, "a: b\nc: d\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "o\n", assertNoErrorString(t)(rn.String()))

	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "c"})
	assert.NoError(t, err)
	assert.Equal(t, "n: o\na: b\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "d\n", assertNoErrorString(t)(rn.String()))

	s = `n: o
a: b
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "o"})
	assert.NoError(t, err)
	assert.Nil(t, rn)
	assert.Equal(t, "n: o\na: b\n", assertNoErrorString(t)(node.String()))

	s = `- a
- b
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "a"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}
	assert.Nil(t, rn)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))

	// should not clear n because it is not empty
	s = `n:
  k: v
a: b
c: d
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "n", IfEmpty: true})
	assert.NoError(t, err)
	assert.Equal(t, "n:\n  k: v\na: b\nc: d\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "", assertNoErrorString(t)(rn.String()))

	// should clear n because it is empty
	s = `n: {}
a: b
c: d
`
	node, err = Parse(s)
	assert.NoError(t, err)
	rn, err = node.Pipe(FieldClearer{Name: "n", IfEmpty: true})
	assert.NoError(t, err)
	assert.Equal(t, "a: b\nc: d\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "{}\n", assertNoErrorString(t)(rn.String()))
}

var s = `n: o
a:
  l: m
  b:
  - f: g
  - c: e
  - h: i
r: s
`

func TestLookup_Fn_create(t *testing.T) {
	// primitive
	node, err := Parse(s)
	assert.NoError(t, err)
	rn, err := node.Pipe(PathGetter{
		Path:   []string{"a", "b", "[c=d]", "t", "f", "[=h]"},
		Create: yaml.ScalarNode,
	})
	assert.NoError(t, err)
	assert.Equal(t, `n: o
a:
  l: m
  b:
  - f: g
  - c: e
  - h: i
  - c: d
    t:
      f:
      - h
r: s
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `h
`, assertNoErrorString(t)(rn.String()))
}

func TestLookup_Fn_create2(t *testing.T) {
	node, err := Parse(s)
	assert.NoError(t, err)
	rn, err := node.Pipe(PathGetter{
		Path:   []string{"a", "b", "[c=d]", "t", "f"},
		Create: yaml.SequenceNode,
	})
	assert.NoError(t, err)
	assert.Equal(t, `n: o
a:
  l: m
  b:
  - f: g
  - c: e
  - h: i
  - c: d
    t:
      f: []
r: s
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `[]
`, assertNoErrorString(t)(rn.String()))
}

func TestLookup_Fn_create3(t *testing.T) {
	node, err := Parse(s)
	assert.NoError(t, err)
	rn, err := node.Pipe(LookupCreate(yaml.MappingNode, "a", "b", "[c=d]", "t"))
	assert.NoError(t, err)
	assert.Equal(t, `n: o
a:
  l: m
  b:
  - f: g
  - c: e
  - h: i
  - c: d
    t: {}
r: s
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `{}
`, assertNoErrorString(t)(rn.String()))
}

func TestLookupCreate_4(t *testing.T) {
	node, err := Parse(`
a: {}
`)
	assert.NoError(t, err)
	rn, err := node.Pipe(
		LookupCreate(yaml.MappingNode, "a", "b", "[c=d]", "t", "f", "[=h]"))

	node.YNode().Style = yaml.FlowStyle
	assert.NoError(t, err)
	assert.Equal(t, "{a: {b: [{c: d, t: {f: [h]}}]}}\n", assertNoErrorString(t)(node.String()))
	assert.Equal(t, "h\n", assertNoErrorString(t)(rn.String()))
}

func TestLookup(t *testing.T) {
	s := `n: o
a:
  l: m
  b:
  - f: g
  - c: e
  - c: d
    t:
      u: v
      f:
      - g
      - h
      - i
    j: k
  - h: i
    p: q
r: s
`
	node, err := Parse(s)
	assert.NoError(t, err)

	// primitive
	rn, err := node.Pipe(Lookup("a", "b", "[c=d]", "t", "f", "[=h]"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `h
`, assertNoErrorString(t)(rn.String()))

	// seq
	rn, err = node.Pipe(Lookup("a", "b", "[c=d]", "t", "f"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `- g
- h
- i
`, assertNoErrorString(t)(rn.String()))

	rn, err = node.Pipe(Lookup("a", "b", "[c=d]", "t"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `u: v
f:
- g
- h
- i
`, assertNoErrorString(t)(rn.String()))

	rn, err = node.Pipe(Lookup("a", "b", "[c=d]"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `c: d
t:
  u: v
  f:
  - g
  - h
  - i
j: k
`, assertNoErrorString(t)(rn.String()))

	rn, err = node.Pipe(Lookup("a", "b"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `- f: g
- c: e
- c: d
  t:
    u: v
    f:
    - g
    - h
    - i
  j: k
- h: i
  p: q
`, assertNoErrorString(t)(rn.String()))

	rn, err = node.Pipe(Lookup("l"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("zzz"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("[a=b]"))
	assert.Error(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("a", "b", "f"))
	assert.Error(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("a", "b", "c=zzz"))
	assert.Error(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup(" ", "a", "", "b", " ", "[c=d]", "\n", "t", "\t", "f", "  ", "[=h]", "  "))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `h
`, assertNoErrorString(t)(rn.String()))

	rn, err = node.Pipe(Lookup(" ", "a", "", "b", " ", "[]"))
	assert.Error(t, err)
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("a", "b", "[c=d]", "t", "f", "[=c]"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)

	rn, err = node.Pipe(Lookup("a", "b", "[z=z]"))
	assert.NoError(t, err)
	assert.Equal(t, s, assertNoErrorString(t)(node.String()))
	assert.Nil(t, rn)
}

func TestSetField_Fn(t *testing.T) {
	// Change field
	node, err := Parse(`
foo: baz
`)
	assert.NoError(t, err)
	instance := FieldSetter{
		Name:  "foo",
		Value: NewScalarRNode("bar"),
	}
	k, err := instance.Filter(node)
	assert.NoError(t, err)
	assert.Equal(t, `foo: bar
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `bar
`, assertNoErrorString(t)(k.String()))

	// Add field
	node, err = Parse(`
foo: baz
`)
	assert.NoError(t, err)
	instance = FieldSetter{
		Name:  "bar",
		Value: NewScalarRNode("buz"),
	}
	k, err = instance.Filter(node)
	assert.NoError(t, err)
	assert.Equal(t, `foo: baz
bar: buz
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `buz
`, assertNoErrorString(t)(k.String()))

	// Clear field
	node, err = Parse(`
foo: baz
bar: buz
`)
	assert.NoError(t, err)
	instance = FieldSetter{
		Name: "foo",
	}
	k, err = instance.Filter(node)
	assert.NoError(t, err)
	assert.Equal(t, `bar: buz
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `baz
`, assertNoErrorString(t)(k.String()))

	// Encounter error
	node, err = Parse(`
-a
-b
`)
	assert.NoError(t, err)
	instance = FieldSetter{
		Name:  "foo",
		Value: NewScalarRNode("v"),
	}
	k, err = instance.Filter(node)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}
	assert.Nil(t, k)
}

func TestSet_Fn(t *testing.T) {
	node, err := Parse(`
foo: baz
`)
	assert.NoError(t, err)
	k, err := node.Pipe(Get("foo"), Set(NewScalarRNode("bar")))
	assert.NoError(t, err)
	assert.Equal(t, `foo: bar
`, assertNoErrorString(t)(node.String()))
	assert.Equal(t, `bar
`, assertNoErrorString(t)(k.String()))

	node, err = Parse(`
foo: baz
`)
	assert.NoError(t, err)
	_, err = node.Pipe(Set(NewScalarRNode("bar")))
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "wrong Node Kind")
	assert.Equal(t, `foo: baz
`, assertNoErrorString(t)(node.String()))
}

func TestErrorIfInvalid(t *testing.T) {
	err := ErrorIfInvalid(
		NewRNode(&yaml.Node{Kind: yaml.SequenceNode}), yaml.SequenceNode)
	assert.NoError(t, err)

	// nil values should pass validation -- they were not specified
	err = ErrorIfInvalid(&RNode{}, yaml.SequenceNode)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = ErrorIfInvalid(NewRNode(&Node{Content: []*yaml.Node{{Value: "hello"}}}), yaml.SequenceNode)
	if !assert.Error(t, err) {
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "wrong Node Kind")

	err = ErrorIfInvalid(NewRNode(&yaml.Node{}), yaml.SequenceNode)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}
	err = ErrorIfInvalid(NewRNode(&yaml.Node{}), yaml.MappingNode)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}

	err = ErrorIfInvalid(NewRNode(&yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{{}, {}},
	}), yaml.MappingNode)
	assert.NoError(t, err)

	err = ErrorIfInvalid(NewRNode(&yaml.Node{}), yaml.SequenceNode)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "wrong Node Kind")
	}

	err = ErrorIfInvalid(NewRNode(&yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{{}},
	}), yaml.MappingNode)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "even length")
	}
}

func TestSplitIndexNameValue(t *testing.T) {
	k, v, err := SplitIndexNameValue("")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "fieldName=fieldValue")
	}
	assert.Equal(t, "", k)
	assert.Equal(t, "", v)

	k, v, err = SplitIndexNameValue("a=b")
	assert.NoError(t, err)
	assert.Equal(t, "a", k)
	assert.Equal(t, "b", v)

	k, v, err = SplitIndexNameValue("=b")
	assert.NoError(t, err)
	assert.Equal(t, "", k)
	assert.Equal(t, "b", v)

	k, v, err = SplitIndexNameValue("a=b=c")
	assert.NoError(t, err)
	assert.Equal(t, "a", k)
	assert.Equal(t, "b=c", v)
}

type filter struct {
	fn func(object *RNode) (*RNode, error)
}

func (c filter) Filter(object *RNode) (*RNode, error) {
	return c.fn(object)
}

func TestResourceNode_Pipe(t *testing.T) {
	r0, r1, r2, r3 := &RNode{}, &RNode{}, &RNode{}, &RNode{}
	var called []string

	// check the nil value case
	r0 = nil
	_, err := r0.Pipe(FieldMatcher{Name: "foo"})
	assert.NoError(t, err)

	r0 = &RNode{}
	// all filters successful
	v, err := r0.Pipe(
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, r0 == object)
			called = append(called, "a")
			return r1, nil
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, object == r1, "function arg doesn't match last function output")
			called = append(called, "b")
			return r2, nil
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, object == r2, "function arg doesn't match last function output")
			return r3, nil
		}},
	)
	assert.True(t, v == r3, "expected r3")
	assert.Nil(t, err)
	assert.Equal(t, called, []string{"a", "b"})

	// filter returns nil
	called = []string{}
	v, err = r0.Pipe(
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, r0 == object)
			called = append(called, "a")
			return r1, nil
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, object == r1, "function arg doesn't match last function output")
			called = append(called, "b")
			return nil, nil
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.Fail(t, "function should be run after error")
			return nil, nil
		}},
	)
	assert.Nil(t, v)
	assert.Nil(t, err)
	assert.Equal(t, called, []string{"a", "b"})

	// filter returns an error
	called = []string{}
	v, err = r0.Pipe(
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, r0 == object)
			called = append(called, "a")
			return r1, nil
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.True(t, object == r1, "function arg doesn't match last function output")
			called = append(called, "b")
			return r1, fmt.Errorf("expected-error")
		}},
		filter{fn: func(object *RNode) (*RNode, error) {
			assert.Fail(t, "function should be run after error")
			return nil, nil
		}},
	)
	assert.True(t, v == r1, "expected r1 as value")
	assert.EqualError(t, err, "expected-error")
	assert.Equal(t, called, []string{"a", "b"})
}

func TestClearAnnotation(t *testing.T) {
	// create metadata.annotations field
	r0 := assertNoError(t)(Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
 annotations:
   z: y
   a.b.c: d.e.f
   s: t
`))

	rn := assertNoError(t)(r0.Pipe(ClearAnnotation("a.b.c")))
	assert.Equal(t, "d.e.f\n", assertNoErrorString(t)(rn.String()))
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    z: y
    s: t
`, assertNoErrorString(t)(r0.String()))
}

func TestGetAnnotation(t *testing.T) {
	r0 := assertNoError(t)(Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
 labels:
   app: java
 annotations:
   a.b.c: d.e.f
   g: h
   i: j
   k: l
 name: app`))

	rn := assertNoError(t)(
		r0.Pipe(GetAnnotation("a.b.c")))
	assert.Equal(t, "d.e.f\n", assertNoErrorString(t)(rn.String()))
}

func TestSetAnnotation_Fn(t *testing.T) {
	// create metadata.annotations field
	r0 := assertNoError(t)(Parse(`apiVersion: apps/v1
kind: Deployment`))

	rn := assertNoError(t)(r0.Pipe(SetAnnotation("a.b.c", "d.e.f")))
	assert.Equal(t, "d.e.f\n", assertNoErrorString(t)(rn.String()))
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    a.b.c: d.e.f
`, assertNoErrorString(t)(r0.String()))
}

func TestRNode_GetMeta(t *testing.T) {
	s := `apiVersion: v1/apps
kind: Deployment
metadata:
  name: foo
  namespace: bar
  labels:
    kl: vl
  annotations:
    ka: va
`
	node, err := Parse(s)
	if !assert.NoError(t, err) {
		return
	}
	meta, err := node.GetMeta()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, ResourceMeta{
		Kind:       "Deployment",
		APIVersion: "v1/apps",
		ObjectMeta: ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Annotations: map[string]string{
				"ka": "va",
			},
			Labels: map[string]string{
				"kl": "vl",
			},
		},
	}, meta)
}

func assertNoError(t *testing.T) func(o *RNode, err error) *RNode {
	return func(o *RNode, err error) *RNode {
		assert.NoError(t, err)
		return o
	}
}

func assertNoErrorString(t *testing.T) func(string, error) string {
	return func(s string, err error) string {
		assert.NoError(t, err)
		return s
	}
}
