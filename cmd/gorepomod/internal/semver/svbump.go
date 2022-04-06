// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package semver

type SvBump int

const (
	Patch SvBump = iota
	Minor
	Major
)

func (b SvBump) String() string {
	return map[SvBump]string{
		Patch: "Patch",
		Minor: "Minor",
		Major: "Major",
	}[b]
}
