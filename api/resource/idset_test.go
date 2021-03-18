// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/api/resource"
)

func TestIdSet_Empty(t *testing.T) {
	s := MakeIdSet([]*Resource{})
	assert.Equal(t, 0, s.Size())
	assert.False(t, s.Contains(testDeployment.CurId()))
	assert.False(t, s.Contains(testConfigMap.CurId()))
}

func TestIdSet_One(t *testing.T) {
	s := MakeIdSet([]*Resource{testDeployment})
	assert.Equal(t, 1, s.Size())
	assert.True(t, s.Contains(testDeployment.CurId()))
	assert.False(t, s.Contains(testConfigMap.CurId()))
}

func TestIdSet_Two(t *testing.T) {
	s := MakeIdSet([]*Resource{testDeployment, testConfigMap})
	assert.Equal(t, 2, s.Size())
	assert.True(t, s.Contains(testDeployment.CurId()))
	assert.True(t, s.Contains(testConfigMap.CurId()))
}
