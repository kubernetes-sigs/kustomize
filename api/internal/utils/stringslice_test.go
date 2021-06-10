// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/internal/utils"
)

func TestStringSliceIndex(t *testing.T) {
	assert.Equal(t, 0, StringSliceIndex([]string{"a", "b"}, "a"))
	assert.Equal(t, 1, StringSliceIndex([]string{"a", "b"}, "b"))
	assert.Equal(t, -1, StringSliceIndex([]string{"a", "b"}, "c"))
	assert.Equal(t, -1, StringSliceIndex([]string{}, "c"))
}

func TestStringSliceContains(t *testing.T) {
	assert.True(t, StringSliceContains([]string{"a", "b"}, "a"))
	assert.True(t, StringSliceContains([]string{"a", "b"}, "b"))
	assert.False(t, StringSliceContains([]string{"a", "b"}, "c"))
	assert.False(t, StringSliceContains([]string{}, "c"))
}

func TestSameEndingSubarray(t *testing.T) {
	assert.True(t, SameEndingSubSlice([]string{"", "a", "b"}, []string{"a", "b"}))
	assert.True(t, SameEndingSubSlice([]string{"a", "b", ""}, []string{"b", ""}))
	assert.True(t, SameEndingSubSlice([]string{"a", "b"}, []string{"a", "b"}))
	assert.True(t, SameEndingSubSlice([]string{"a", "b"}, []string{"b"}))
	assert.True(t, SameEndingSubSlice([]string{"b"}, []string{"a", "b"}))
	assert.True(t, SameEndingSubSlice([]string{}, []string{}))
	assert.False(t, SameEndingSubSlice([]string{"a", "b"}, []string{"b", "a"}))
	assert.False(t, SameEndingSubSlice([]string{"a", "b"}, []string{}))
	assert.False(t, SameEndingSubSlice([]string{"a", "b"}, []string{""}))
}
