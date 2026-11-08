// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/resource"
)

// Origin annotations record paths with "/" on every platform, like the
// paths in a kustomization file, so build output does not depend on the
// OS kustomize runs on.
func TestOriginAppend_Windows(t *testing.T) {
	actual, err := (&Origin{Path: `overlay\prod`}).Append(`..\base\service.yaml`).String()
	require.NoError(t, err)
	assert.Equal(t, "path: overlay/base/service.yaml\n", actual)
}
