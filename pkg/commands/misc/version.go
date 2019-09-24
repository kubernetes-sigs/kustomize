// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package misc

import (
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	kustomizeVersion = "unknown"
	goos             = runtime.GOOS
	goarch           = runtime.GOARCH
	gitCommit        = "$Format:%H$" // sha1 from git, output of $(git rev-parse HEAD)

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// version returns the version of kustomize.
type version struct {
	// KustomizeVersion is a kustomize binary version.
	KustomizeVersion string `json:"kustomizeVersion"`
	// GitCommit is a git commit
	GitCommit string `json:"gitCommit"`
	// BuildDate is a build date of the binary.
	BuildDate string `json:"buildDate"`
	// GoOs holds OS name.
	GoOs string `json:"goOs"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch"`
}

// getVersion returns version.
func getVersion() version {
	return version{
		kustomizeVersion,
		gitCommit,
		buildDate,
		goos,
		goarch,
	}
}

// Print prints version.
func (v version) Print(w io.Writer, short bool) {
	if short {
		fmt.Fprintf(w, "%s\n", v.KustomizeVersion)
	} else {
		fmt.Fprintf(w, "Version: %+v\n", v)
	}

}

// NewCmdVersion makes version command.
func NewCmdVersion(w io.Writer) *cobra.Command {
	var short bool

	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Prints the kustomize version",
		Example: `kustomize version`,
		Run: func(cmd *cobra.Command, args []string) {
			getVersion().Print(w, short)
		},
	}

	versionCmd.Flags().BoolVar(&short, "short", false, "print just the version number")
	return &versionCmd
}
