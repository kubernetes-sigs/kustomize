// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio"
)

func TestPipe(t *testing.T) {
	p := Pipeline{
		Inputs:  []Reader{},
		Filters: []Filter{},
		Outputs: []Writer{},
	}

	err := p.Execute()
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
}

func TestSlice_Write(t *testing.T) {

}
