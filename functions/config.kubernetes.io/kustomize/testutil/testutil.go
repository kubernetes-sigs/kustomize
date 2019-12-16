// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

type FieldSpecTestCase struct {
	Name     string
	Input    string
	Expected string
	Instance kio.Filter
}

func RunTestCases(t *testing.T, tcs []FieldSpecTestCase) {
	for _, tc := range tcs {
		in := bytes.NewBufferString(tc.Input)
		out := &bytes.Buffer{}
		rw := &kio.ByteReadWriter{Reader: in, Writer: out}
		err := kio.Pipeline{
			Inputs:  []kio.Reader{rw},
			Filters: []kio.Filter{tc.Instance},
			Outputs: []kio.Writer{rw},
		}.Execute()
		if !assert.NoError(t, err, tc.Name) {
			t.FailNow()
		}
		assert.Equal(t, strings.TrimSpace(tc.Expected), strings.TrimSpace(out.String()), tc.Name)
	}
}
