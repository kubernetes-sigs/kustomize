// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldreference_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

type fieldReferenceTestCase struct {
	name     string
	input    string
	expected string
	instance kio.Filter
}

func doTestCases(t *testing.T, tcs []fieldReferenceTestCase) {
	for _, tc := range tcs {
		in := bytes.NewBufferString(tc.input)
		out := &bytes.Buffer{}
		rw := &kio.ByteReadWriter{Reader: in, Writer: out}
		err := kio.Pipeline{
			Inputs:  []kio.Reader{rw},
			Filters: []kio.Filter{tc.instance},
			Outputs: []kio.Writer{rw},
		}.Execute()
		if !assert.NoError(t, err, tc.name) {
			t.FailNow()
		}
		assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(out.String()), tc.name)
	}
}
