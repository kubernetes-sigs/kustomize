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
		Path:    provenance.ModulePath,
		Version: version,
		Replace: nil,
	}
}

func mockBuildInfo(mainVersion, depsVersion string) *debug.BuildInfo {
	module := mockModule(depsVersion)

	return &debug.BuildInfo{
		Main: debug.Module{
			Version: mainVersion,
		},
		Deps: []*debug.Module{
			&module,
		},
	}
}

func TestFindVersion(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		buildInfo       *debug.BuildInfo
		expectedVersion string
	}{
		{
			name:            "The version from LD_FLAGS is not overridden by main and dependencies versions",
			version:         "v2.3.4",
			buildInfo:       mockBuildInfo("v1.2.3", "(devel)"),
			expectedVersion: "v2.3.4",
		},
		{
			name:            "The version from LD_FLAGS is overridden by the main version",
			version:         "(devel)",
			buildInfo:       mockBuildInfo("v1.2.3", "(devel)"),
			expectedVersion: "v1.2.3",
		},
		{
			name:            "The version from LD_FLAGS is overridden by the version from dependencies",
			version:         "(devel)",
			buildInfo:       mockBuildInfo("v1.2.3", "v1.2.3-0.20210101010101-abcdefabcdef"),
			expectedVersion: "v1.2.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := provenance.FindVersion(tt.buildInfo, tt.version)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestGetMostRecentTag(t *testing.T) {
	tests := []struct {
		name        string
		module      debug.Module
		isError     bool
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
			name:    "Invalid semver string",
			module:  mockModule("invalid-version"),
			isError: true,
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
			tag, err := provenance.GetMostRecentTag(tt.module)
			if err != nil {
				if !tt.isError {
					assert.NoError(t, err)
				}
			} else {
				assert.Equal(t, tt.expectedTag, tag)
			}
		})
	}
}
