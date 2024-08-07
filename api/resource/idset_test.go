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
	td, err := createTestDeployment()
	if err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}
	tc, err := createTestConfigMap()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	assert.Equal(t, 0, s.Size())
	assert.False(t, s.Contains(td.CurId()))
	assert.False(t, s.Contains(tc.CurId()))
}

func TestIdSet_One(t *testing.T) {
	td, err := createTestDeployment()
	if err != nil {
		t.Fatalf("failed to create test deployment: %v", err)
	}
	tc, err := createTestConfigMap()
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}
	s := MakeIdSet([]*Resource{td})
	assert.Equal(t, 1, s.Size())
	assert.True(t, s.Contains(td.CurId()))
	assert.False(t, s.Contains(tc.CurId()))
}

func TestIdSet_Two(t *testing.T) {
	td, err := createTestDeployment()
	if err != nil {
		t.Fatalf("failed to create test Deployment: %v", err)
	}
	tc, err := createTestConfigMap()
	if err != nil {
		t.Fatalf("failed to create test Config: %v", err)
	}
	s := MakeIdSet([]*Resource{td, tc})
	assert.Equal(t, 2, s.Size())
	assert.True(t, s.Contains(td.CurId()))
	assert.True(t, s.Contains(tc.CurId()))
}
