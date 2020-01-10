// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that non-UTF8 characters in comments don't cause failures
func TestRNode_GetMeta_UTF16(t *testing.T) {
	sr, err := Parse(`apiVersion: rbac.istio.io/v1alpha1
kind: ServiceRole
metadata:
  name: wildcard
  namespace: default
  # If set to [“*”], it refers to all services in the namespace
  annotations:
    foo: bar
spec:
  rules:
    # There is one service in default namespace, should not result in a validation error
    # If set to [“*”], it refers to all services in the namespace
    - services: ["*"]
      methods: ["GET", "HEAD"]
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	actual, err := sr.GetMeta()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := ResourceMeta{
		APIVersion: "rbac.istio.io/v1alpha1",
		Kind:       "ServiceRole",
		ObjectMeta: ObjectMeta{
			Name:        "wildcard",
			Namespace:   "default",
			Annotations: map[string]string{"foo": "bar"},
		},
	}
	if !assert.Equal(t, expected, actual) {
		t.FailNow()
	}
}
