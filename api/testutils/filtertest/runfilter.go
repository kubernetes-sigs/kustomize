// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filtertest_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func run(input string, f kio.Filter) (string, error) {
	var out bytes.Buffer
	rw := kio.ByteReadWriter{
		Reader: bytes.NewBufferString(input),
		Writer: &out,
	}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{&rw},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{&rw},
	}.Execute()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// RunFilter runs filter and panic if there is error
func RunFilter(t *testing.T, input string, f kio.Filter) string {
	output, err := run(input, f)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return output
}

// RunFilterE runs filter and return error if there is
func RunFilterE(t *testing.T, input string, f kio.Filter) (string, error) {
	output, err := run(input, f)
	if err != nil {
		return "", err
	}
	return output, nil
}
