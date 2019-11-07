// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"fmt"
	"log"

	. "sigs.k8s.io/kustomize/kyaml/yaml"
)

func Example() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
`)
	if err != nil {
		log.Fatal(err)
	}

	containers, err := Parse(`
- name: nginx # first container
  image: nginx
- name: nginx2 # second container
  image: nginx2
`)
	if err != nil {
		log.Fatal(err)
	}

	node, err := obj.Pipe(
		LookupCreate(SequenceNode, "spec", "template", "spec", "containers"),
		Append(containers.YNode().Content...))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	//  <nil>
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: app
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx # first container
	//         image: nginx
	//       - name: nginx2 # second container
	//         image: nginx2
	//  <nil>
}

func ExampleAppend_appendScalars() {
	obj, err := Parse(`
- a
- b
`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = obj.Pipe(Append(&Node{Value: "c", Kind: ScalarNode}))
	if err != nil {
		log.Fatal(err)
	}
	node, err := obj.Pipe(Append(
		&Node{Value: "c", Kind: ScalarNode},
		&Node{Value: "d", Kind: ScalarNode},
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	//  <nil>
	// - a
	// - b
	// - c
	// - c
	// - d
	//  <nil>
}

func ExampleAppend_appendMap() {
	obj, err := Parse(`
- name: foo
- name: bar
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := Parse("name: baz")
	if err != nil {
		log.Fatal(err)
	}
	node, err := obj.Pipe(Append(elem.YNode()))
	if err != nil {
		log.Fatal(err)
	}

	// Expect the node to contain the appended element because only
	// 1 element was appended
	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	// name: baz
	//  <nil>
	// - name: foo
	// - name: bar
	// - name: baz
	//  <nil>
}

func ExampleClear() {
	obj, err := Parse(`
kind: Deployment
metadata:
  name: app
  annotations:
    a.b.c: d.e.f
    g: h
spec:
  template: {}
`)
	if err != nil {
		log.Fatal(err)
	}
	node, err := obj.Pipe(Clear("metadata"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	// name: app
	// annotations:
	//   a.b.c: d.e.f
	//   g: h
	//  <nil>
	// kind: Deployment
	// spec:
	//   template: {}
	//  <nil>
}

func ExampleGet() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  annotations:
    a.b.c: d.e.f
    g: h
spec:
  template: {}
`)
	if err != nil {
		log.Fatal(err)
	}
	node, err := obj.Pipe(Get("metadata"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	// name: app
	// annotations:
	//   a.b.c: d.e.f
	//   g: h
	//  <nil>
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: app
	//   annotations:
	//     a.b.c: d.e.f
	//     g: h
	// spec:
	//   template: {}
	//  <nil>
}

func ExampleGet_notFound() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
spec:
  template: {}
`)
	if err != nil {
		log.Fatal(err)
	}
	node, err := obj.Pipe(FieldMatcher{Name: "metadata"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node.String())
	fmt.Println(obj.String())
	// Output:
	//  <nil>
	// apiVersion: apps/v1
	// kind: Deployment
	// spec:
	//   template: {}
	//  <nil>
}

func ExampleElementMatcher_Filter() {
	obj, err := Parse(`
- a
- b
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := obj.Pipe(ElementMatcher{
		FieldValue: "c", Create: NewScalarRNode("c"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(elem.String())
	fmt.Println(obj.String())
	// Output:
	// c
	//  <nil>
	// - a
	// - b
	// - c
	//  <nil>
}

func ExampleElementMatcher_Filter_primitiveFound() {
	obj, err := Parse(`
- a
- b
- c
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := obj.Pipe(ElementMatcher{
		FieldValue: "c", Create: NewScalarRNode("c"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(elem.String())
	fmt.Println(obj.String())
	// Output:
	// c
	//  <nil>
	// - a
	// - b
	// - c
	//  <nil>
}

func ExampleElementMatcher_Filter_objectNotFound() {
	obj, err := Parse(`
- name: foo
- name: bar
`)
	if err != nil {
		log.Fatal(err)
	}
	append, err := Parse(`
name: baz
image: nginx
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := obj.Pipe(ElementMatcher{
		FieldName: "name", FieldValue: "baz", Create: append})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(elem.String())
	fmt.Println(obj.String())
	// Output:
	// name: baz
	// image: nginx
	//  <nil>
	// - name: foo
	// - name: bar
	// - name: baz
	//   image: nginx
	//  <nil>
}

func ExampleElementMatcher_Filter_objectFound() {
	obj, err := Parse(`
- name: foo
- name: bar
- name: baz
`)
	if err != nil {
		log.Fatal(err)
	}
	append, err := Parse(`
name: baz
image: nginx
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := obj.Pipe(ElementMatcher{
		FieldName: "name", FieldValue: "baz", Create: append})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(elem.String())
	fmt.Println(obj.String())
	// Output:
	// name: baz
	//  <nil>
	// - name: foo
	// - name: bar
	// - name: baz
	//  <nil>
}

func ExampleFieldMatcher_Filter() {
	obj, err := Parse(`
kind: Deployment
spec:
  template: {}
`)
	if err != nil {
		log.Fatal(err)
	}
	value, err := Parse(`
name: app
annotations:
  a.b.c: d.e.f
  g: h
`)
	if err != nil {
		log.Fatal(err)
	}
	elem, err := obj.Pipe(FieldMatcher{
		Name: "metadata", Value: value, Create: value})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(elem.String())
	fmt.Println(obj.String())
	// Output:
	// name: app
	// annotations:
	//   a.b.c: d.e.f
	//   g: h
	//  <nil>
	// kind: Deployment
	// spec:
	//   template: {}
	// metadata:
	//   name: app
	//   annotations:
	//     a.b.c: d.e.f
	//     g: h
	//  <nil>
}

func ExampleLookup_element() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
spec:
  templates:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`)
	if err != nil {
		log.Fatal(err)
	}
	value, err := obj.Pipe(Lookup(
		"spec", "templates", "spec", "containers", "[name=nginx]"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value.String())
	// Output:
	// name: nginx
	// image: nginx:latest
	//  <nil>
}

func ExampleLookup_sequence() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
     app: java
  annotations:
    a.b.c: d.e.f
  name: app
spec:
  templates:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`)
	if err != nil {
		log.Fatal(err)
	}
	value, err := obj.Pipe(Lookup(
		"spec", "templates", "spec", "containers"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value.String())
	// Output:
	// - name: nginx
	//   image: nginx:latest
	//  <nil>
}

func ExampleLookup_scalar() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
spec:
  templates:
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`)
	if err != nil {
		log.Fatal(err)
	}
	value, err := obj.Pipe(Lookup(
		"spec", "templates", "spec", "containers", "[name=nginx]", "image"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value.String())
	// Output:
	// nginx:latest
	//  <nil>
}

func ExampleLookupCreate_element() {
	obj, err := Parse(`
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := obj.Pipe(LookupCreate(
		MappingNode, "spec", "templates", "spec", "containers", "[name=nginx]"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rs.String())
	fmt.Println("---")
	fmt.Println(obj.String())
	// Output:
	// name: nginx
	//  <nil>
	// ---
	// kind: Deployment
	// metadata:
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	//   name: app
	// spec:
	//   templates:
	//     spec:
	//       containers:
	//       - name: nginx
	//  <nil>
}

func ExampleLookupCreate_sequence() {
	obj, err := Parse(`
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := obj.Pipe(LookupCreate(
		SequenceNode, "spec", "templates", "spec", "containers"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rs.String())
	fmt.Println("---")
	fmt.Println(obj.String())
	// Output:
	// []
	//  <nil>
	// ---
	// kind: Deployment
	// metadata:
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	//   name: app
	// spec:
	//   templates:
	//     spec:
	//       containers: []
	//  <nil>
}

func ExamplePathGetter_Filter() {
	obj, err := Parse(`
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := obj.Pipe(PathGetter{
		Path:   []string{"spec", "templates", "spec", "containers", "[name=nginx]", "image"},
		Create: ScalarNode,
	})
	if err != nil {
		log.Fatal(err)
	}
	rs.Document().Style = SingleQuotedStyle

	fmt.Println(rs.String())
	fmt.Println("---")
	fmt.Println(obj.String())
	// Output:
	// ''
	//  <nil>
	// ---
	// kind: Deployment
	// metadata:
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	//   name: app
	// spec:
	//   templates:
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: ''
	//  <nil>
}

func ExampleLookupCreate_object() {
	obj, err := Parse(`
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := obj.Pipe(LookupCreate(
		MappingNode, "spec", "templates", "spec"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rs.String())
	fmt.Println("---")
	fmt.Println(obj.String())
	// Output:
	// {}
	//  <nil>
	// ---
	// kind: Deployment
	// metadata:
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	//   name: app
	// spec:
	//   templates:
	//     spec: {}
	//  <nil>
}

func ExampleLookup_notFound() {
	obj, err := Parse(`
kind: Deployment
metadata:
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
  name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := obj.Pipe(Lookup("spec", "templates", "spec"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rs)
	fmt.Println("---")
	fmt.Println(obj.String())
	// Output:
	//  <nil>
	// ---
	// kind: Deployment
	// metadata:
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	//   name: app
	//  <nil>
}

func ExampleSetField_stringValue() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
 name: app
`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = obj.Pipe(SetField("foo", NewScalarRNode("bar")))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(obj.String())
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: app
	// foo: bar
	//  <nil>
}

func ExampleSetField_stringValueOverwrite() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
   name: app
foo: baz
`)
	if err != nil {
		// handle error
	}
	// set metadata.annotations.foo = bar
	_, err = obj.Pipe(SetField("foo", NewScalarRNode("bar")))
	if err != nil {
		// handle error
	}

	fmt.Println(obj.String())
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: app
	// foo: bar
	//  <nil>
}

func ExampleSetField() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
  annotations:
    a.b.c: d.e.f
`)
	if err != nil {
		log.Fatal(err)
	}

	containers, err := Parse(`
- name: nginx # first container
  image: nginx
- name: nginx2 # second container
  image: nginx2
`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = obj.Pipe(
		LookupCreate(MappingNode, "spec", "template", "spec"),
		SetField("containers", containers))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(obj.String())
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: app
	//   labels:
	//     app: java
	//   annotations:
	//     a.b.c: d.e.f
	// spec:
	//   template:
	//     spec:
	//       containers:
	//       - name: nginx # first container
	//         image: nginx
	//       - name: nginx2 # second container
	//         image: nginx2
	//  <nil>
}

func ExampleTee() {
	obj, err := Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
`)
	if err != nil {
		// handle error
	}
	// set metadata.annotations.foo = bar
	_, err = obj.Pipe(
		Lookup("spec", "template", "spec", "containers", "[name=nginx]"),
		Tee(SetField("filter", NewListRNode("foo"))),
		SetField("args", NewListRNode("baz", "bar")))
	if err != nil {
		// handle error
	}

	fmt.Println(obj.String())
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: nginx-deployment
	//   labels:
	//     app: nginx
	// spec:
	//   replicas: 3
	//   selector:
	//     matchLabels:
	//       app: nginx
	//   template:
	//     metadata:
	//       labels:
	//         app: nginx
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.7.9
	//         ports:
	//         - containerPort: 80
	//         filter:
	//         - foo
	//         args:
	//         - baz
	//         - bar
	//  <nil>
}

func ExampleRNode_Elements() {
	resource, err := Parse(`
- name: foo
  args: ['run.sh']
- name: bar
  args: ['run.sh']
- name: baz
  args: ['run.sh']
`)
	if err != nil {
		log.Fatal(err)
	}
	elements, err := resource.Elements()
	if err != nil {
		log.Fatal(err)
	}
	for i, e := range elements {
		fmt.Println(fmt.Sprintf("Element: %d", i))
		fmt.Println(e.MustString())
	}
	// Output:
	// Element: 0
	// name: foo
	// args: ['run.sh']
	//
	// Element: 1
	// name: bar
	// args: ['run.sh']
	//
	// Element: 2
	// name: baz
	// args: ['run.sh']

}

func ExampleRNode_ElementValues() {
	resource, err := Parse(`
- name: foo
  args: ['run.sh']
- name: bar
  args: ['run.sh']
- name: baz
  args: ['run.sh']
`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resource.ElementValues("name"))
	// Output:
	// [foo bar baz] <nil>
}
