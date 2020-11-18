// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var input = `apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`

func TestSetK8sData(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
`)
	_, err := rn.Pipe(
		SetK8sData("foo", "bar"),
		SetK8sData("fruit", "apple"),
		SetK8sData("veggie", "celery"))
	assert.NoError(t, err)
	output := rn.MustString()

	expected := `apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
  foo: bar
  fruit: apple
  veggie: celery
`
	if !assert.Equal(t, expected, output) {
		t.FailNow()
	}
}

func TestSetK8sDataForbidOverwrite(t *testing.T) {
	rn := MustParse(`apiVersion: v1
kind: ConfigMap
data:
  altGreeting: "Good Morning!"
`)
	_, err := rn.Pipe(
		SetK8sData("foo", "bar"),
		SetK8sData("altGreeting", "hey"),
		SetK8sData("veggie", "celery"))
	assert.EqualError(
		t, err, "protecting existing altGreeting='\"Good Morning!\"' "+
			"against attempt to add new value 'hey'")
}

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
	rn := MustParse(input)
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
	rn := MustParse(input)
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
