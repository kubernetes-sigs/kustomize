// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"bytes"
	"runtime"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"testing"

	goerrors "github.com/go-errors/errors"

	"github.com/stretchr/testify/assert"
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

func AssertErrorContains(t *testing.T, err error, value string, msg ...string) {
	if !assert.Error(t, err, msg) {
		t.FailNow()
	}
	if !assert.Contains(t, err.Error(), value, msg) {
		t.FailNow()
	}
}

func AssertNoError(t *testing.T, err error, msg ...string) {
	if !assert.NoError(t, err, msg) {
		gerr, ok := err.(*goerrors.Error)
		if ok {
			t.Fatal(string(gerr.Stack()))
		}
		t.FailNow()
	}
}

func SkipWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
}
