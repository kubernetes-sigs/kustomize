// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortedKeys(t *testing.T) {
	testCases := map[string]struct {
		input    map[string]string
		expected []string
	}{
		"empty": {
			input:    map[string]string{},
			expected: []string{}},
		"one": {
			input:    map[string]string{"a": "aaa"},
			expected: []string{"a"}},
		"three": {
			input:    map[string]string{"c": "ccc", "b": "bbb", "a": "aaa"},
			expected: []string{"a", "b", "c"}},
	}
	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			if !assert.Equal(t,
				yaml.SortedMapKeys(tc.input),
				tc.expected) {
				t.FailNow()
			}
		})
	}
}
