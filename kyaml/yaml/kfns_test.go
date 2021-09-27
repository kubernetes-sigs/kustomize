// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetMeta(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
`)
	_, err := rn.Pipe(SetK8sName("foo"), SetK8sNamespace("bar"))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	output := rn.MustString()

	expected := `apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
metadata:
  name: foo
  namespace: bar
`
	if !assert.Equal(t, expected, output) {
		t.FailNow()
	}
}

func TestSetLabel1(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`)
	_, err := rn.Pipe(SetLabel("foo", "bar"))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	output := rn.MustString()

	expected := `apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
  labels:
    foo: 'bar'
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`
	if output != expected {
		t.Fatalf("expected \n%s\nbut got \n%s\n", expected, output)
	}
}

func TestSetLabel2(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
`)
	_, err := rn.Pipe(SetLabel("foo", "bar"))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	output := rn.MustString()

	expected := `apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
metadata:
  labels:
    foo: 'bar'
`
	if output != expected {
		t.Fatalf("expected \n%s\nbut got \n%s\n", expected, output)
	}
}

func TestAnnotation(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`)
	_, err := rn.Pipe(SetAnnotation("foo", "bar"))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	output := rn.MustString()

	expected := `apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
  annotations:
    foo: 'bar'
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`
	if output != expected {
		t.Fatalf("expected \n%s\nbut got \n%s\n", expected, output)
	}
}
