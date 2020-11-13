// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package sets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringList_Len(t *testing.T) {
	var sl0 StringList = [][]string{}
	assert.Equal(t, 0, sl0.Len())

	var sl1 StringList = [][]string{
		{""},
	}
	assert.Equal(t, 1, sl1.Len())

	var sl2 StringList = [][]string{
		{"a", "b"},
		{"b", "c"},
	}
	assert.Equal(t, 2, sl2.Len())
}

func TestStringList_Insert(t *testing.T) {
	var sl StringList
	assert.Equal(t, 0, sl.Len())

	sl = sl.Insert([]string{"a", "b", "c"})
	assert.Equal(t, 1, sl.Len())

	sl = sl.Insert([]string{"c", "b", "a"})
	assert.Equal(t, 2, sl.Len())

	sl = sl.Insert([]string{"a", "b", "c"})
	assert.Equal(t, 2, sl.Len())
}

func TestStringList_Has(t *testing.T) {
	var sl StringList

	assert.False(t, sl.Has([]string{}))
	assert.False(t, sl.Has([]string{"a", "b", "c"}))

	sl = sl.Insert([]string{"a", "b", "c"})
	assert.True(t, sl.Has([]string{"a", "b", "c"}))
	assert.False(t, sl.Has([]string{"b", "c", "a"}))

	sl = sl.Insert([]string{})
	assert.True(t, sl.Has([]string{}))
}
