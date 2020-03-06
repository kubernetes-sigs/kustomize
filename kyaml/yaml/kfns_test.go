// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"
)

var input = `apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map
data:
  altGreeting: "Good Morning!"
  enableRisky: "false"
`

func TestSetLabel(t *testing.T) {
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
