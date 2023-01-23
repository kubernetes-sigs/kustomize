// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provenance

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

//nolint:gochecknoglobals
var (
	version = "unknown"
	// sha1 from git, output of $(git rev-parse HEAD)
	gitCommit = "unknown"
	// build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate = "unknown"
)

// Provenance holds information about the build of an executable.
type Provenance struct {
	// Version of the kustomize binary.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// GitCommit is a git commit
	GitCommit string `json:"gitCommit,omitempty" yaml:"gitCommit,omitempty"`
	// BuildDate is date of the build.
	BuildDate string `json:"buildDate,omitempty" yaml:"buildDate,omitempty"`
	// GoOs holds OS name.
	GoOs string `json:"goOs,omitempty" yaml:"goOs,omitempty"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch,omitempty" yaml:"goArch,omitempty"`
	// GoVersion holds Go version.
	GoVersion string `json:"goVersion,omitempty" yaml:"goVersion,omitempty"`
}

// GetProvenance returns an instance of Provenance.
func GetProvenance() Provenance {
	p := Provenance{
		BuildDate: buildDate,
		Version:   version,
		GitCommit: gitCommit,
		GoOs:      runtime.GOOS,
		GoArch:    runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return p
	}

	// override with values from BuildInfo
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if p.Version == "(devel)" {
				p.GitCommit = setting.Value
			}
		case "GOOS":
			p.GoOs = setting.Value
		}
	}
	return p
}

// Full returns the full provenance stamp.
func (v Provenance) Full() string {
	return fmt.Sprintf("%+v", v)
}

// Short returns the semantic version.
func (v Provenance) Short() string {
	return v.Semver()
}

// Semver returns the semantic version of kustomize.
// kustomize version is set in format "kustomize/vX.X.X" in every release.
// X.X.X is a semver. If the version string is not in this format,
// return the original version string
func (v Provenance) Semver() string {
	return strings.TrimPrefix(v.Version, "kustomize/")
}
