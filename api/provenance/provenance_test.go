// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provenance_test

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provenance"
)

const expectedBuildDateFromLdFlag = "2023-01-31T23:38:41Z"
const expectedVersionFromLdFlag = "(test)"

func TestGetProvenance(t *testing.T) {
	p := provenance.GetProvenance()
	// These are set by ldflags in our Makefile
	assert.Equal(t, expectedVersionFromLdFlag, p.Version)
	assert.Equal(t, expectedBuildDateFromLdFlag, p.BuildDate)
	// This comes from BuildInfo, which is not set during go test: https://github.com/golang/go/issues/33976
	assert.Equal(t, "unknown", p.GitCommit)

	// These are set properly during go test
	assert.NotEmpty(t, p.GoArch)
	assert.NotEmpty(t, p.GoOs)
	assert.Contains(t, p.GoVersion, "go1.")
}

func TestProvenance_Short(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test, so this comes from an ldflag: https://github.com/golang/go/issues/33976
	assert.Equal(t, fmt.Sprintf("{%s  %s   }", expectedVersionFromLdFlag, expectedBuildDateFromLdFlag), p.Short())

	p.Version = "kustomize/v4.11.12"
	assert.Equal(t, fmt.Sprintf("{kustomize/v4.11.12  %s   }", expectedBuildDateFromLdFlag), p.Short())
}

func TestProvenance_Semver(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test
	assert.Equal(t, "(test)", p.Semver())

	p.Version = "kustomize/v4.11.12"
	assert.Equal(t, "v4.11.12", p.Semver())
}

func mockModule(version string) debug.Module {
	return debug.Module{
		Path:    "sigs.k8s.io/kustomize/kustomize/v5",
		Version: version,
		Replace: nil,
	}
}

func TestGetMostRecentTag(t *testing.T) {
	tests := []struct {
		name        string
		module      debug.Module
		expectedTag string
	}{
		{
			name:        "Standard version",
			module:      mockModule("v1.2.3"),
			expectedTag: "v1.2.3",
		},
		{
			name:        "Pseudo-version with patch increment",
			module:      mockModule("v0.0.0-20210101010101-abcdefabcdef"),
			expectedTag: "v0.0.0",
		},
		{
			name:        "Invalid semver string",
			module:      mockModule("invalid-version"),
			expectedTag: "unknown",
		},
		{
			name:        "Valid semver with patch increment and pre-release info",
			module:      mockModule("v1.2.3-0.20210101010101-abcdefabcdef"),
			expectedTag: "v1.2.2",
		},
		{
			name:        "Valid semver no patch increment",
			module:      mockModule("v1.2.0"),
			expectedTag: "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := provenance.GetMostRecentTag(tt.module)
			assert.Equal(t, tt.expectedTag, tag)
		})
	}
}
