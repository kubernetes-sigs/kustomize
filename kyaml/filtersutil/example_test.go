// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filtersutil_test

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplyToJSON(t *testing.T) {
	// testFilter sets `foo: bar` on each resource
	testFilter := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			set := yaml.SetField(
				"foo", yaml.NewScalarRNode("bar"))
			err := node.PipeE(set)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			return node, nil
		}))

	obj1 := buffer{Buffer: bytes.NewBufferString(`{"kind": "Foo"}`)}
	obj2 := buffer{Buffer: bytes.NewBufferString(`{"kind": "Bar"}`)}
	err := filtersutil.ApplyToJSON(testFilter, obj1, obj2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(obj1.String())
	fmt.Println(obj2.String())

	// Output:
	// {"foo":"bar","kind":"Foo"}
	// {"foo":"bar","kind":"Bar"}
}

type buffer struct {
	*bytes.Buffer
}

func (buff buffer) UnmarshalJSON(b []byte) error {
	buff.Reset()
	buff.Write(b)
	return nil
}

func (buff buffer) MarshalJSON() ([]byte, error) {
	return buff.Bytes(), nil
}
