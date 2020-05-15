// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil_test

import (
	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func UpdateYamlString(doc string, functions ...yaml.Filter) (string, error) {
	b, err := UpdateYamlBytes([]byte(doc), functions...)
	return string(b), err
}

func UpdateYamlBytes(b []byte, function ...yaml.Filter) ([]byte, error) {
	var out bytes.Buffer
	rw := kio.ByteReadWriter{
		Reader: bytes.NewBuffer(b),
		Writer: &out,
	}
	err := kio.Pipeline{
		Inputs: []kio.Reader{&rw},
		Filters: []kio.Filter{
			kio.FilterAll(yaml.FilterFunc(
				func(node *yaml.RNode) (*yaml.RNode, error) {
					return node.Pipe(function...)
				}),
			),
		},
		Outputs: []kio.Writer{&rw},
	}.Execute()
	return out.Bytes(), err
}
