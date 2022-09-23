// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer //nolint:testpackage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlBase(t *testing.T) {
	require.Equal(t, "repo", urlBase("https://github.com/org/repo"))
}

func TestUrlBaseTrailingSlash(t *testing.T) {
	require.Equal(t, "repo", urlBase("github.com/org/repo//"))
}
