// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"foo", "bar"}, "bar"))
	assert.False(t, Contains([]string{"foo", "bar"}, "baz"))
	assert.False(t, Contains([]string{}, "bar"))
	assert.False(t, Contains([]string{}, ""))
}

func TestRemove(t *testing.T) {
	assert.Equal(t, Remove([]string{"foo", "bar"}, "bar"), []string{"foo"})
	assert.Equal(t, Remove([]string{"foo", "bar", "foo"}, "foo"), []string{"bar", "foo"})
	assert.Equal(t, Remove([]string{"foo"}, "foo"), []string{})
	assert.Equal(t, Remove([]string{}, "foo"), []string{})
	assert.Equal(t, Remove([]string{"foo", "bar", "foo"}, "baz"), []string{"foo", "bar", "foo"})
}
