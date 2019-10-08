// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provenance

import (
	"fmt"
	"io"
	"runtime"
)

var (
	version = "unknown"
	// sha1 from git, output of $(git rev-parse HEAD)
	gitCommit = "$Format:%H$"
	// build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate = "1970-01-01T00:00:00Z"
	goos      = runtime.GOOS
	goarch    = runtime.GOARCH
)

// Provenance holds information about the build of an executable.
type Provenance struct {
	// Version of the kustomize binary.
	Version string `json:"version"`
	// GitCommit is a git commit
	GitCommit string `json:"gitCommit"`
	// BuildDate is date of the build.
	BuildDate string `json:"buildDate"`
	// GoOs holds OS name.
	GoOs string `json:"goOs"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch"`
}

// GetProvenance returns an instance of Provenance.
func GetProvenance() Provenance {
	return Provenance{
		version,
		gitCommit,
		buildDate,
		goos,
		goarch,
	}
}

// Print prints provenance info.
func (v Provenance) Print(w io.Writer, short bool) {
	if short {
		fmt.Fprintf(w, "%s\n", v.Version)
	} else {
		fmt.Fprintf(w, "Version: %+v\n", v)
	}
}
