// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filtertest_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func RunFilter(t *testing.T, input string, f kio.Filter) string {
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
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return out.String()
}
