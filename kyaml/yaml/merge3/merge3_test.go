// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/yaml/merge3"
)

var testCases = [][]testCase{scalarTestCases, listTestCases, mapTestCases, elementTestCases}

func TestMerge(t *testing.T) {
	for i := range testCases {
		for _, tc := range testCases[i] {
			actual, err := MergeStrings(tc.local, tc.origin, tc.update)
			if tc.err == nil {
				if !assert.NoError(t, err, tc.description) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(tc.expected), strings.TrimSpace(actual), tc.description) {
					t.FailNow()
				}
			} else {
				if !assert.Errorf(t, err, tc.description) {
					t.FailNow()
				}
				if !assert.Contains(t, tc.err.Error(), err.Error()) {
					t.FailNow()
				}
			}
		}
	}
}

type testCase struct {
	description string
	origin      string
	update      string
	local       string
	expected    string
	err         error
}
