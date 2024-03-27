// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer is the immutable semantic version per https://semver.org
type SemVer struct {
	major int
	minor int
	patch int
}

func New(major, minor, patch int) SemVer {
	return SemVer{
		major: major,
		minor: minor,
		patch: patch,
	}
}

var zero = New(0, 0, 0)

func Zero() SemVer {
	return zero
}

// Versions implements sort.Interface to get decreasing order.
type Versions []SemVer

func (v Versions) Len() int           { return len(v) }
func (v Versions) Less(i, j int) bool { return v[j].LessThan(v[i]) }
func (v Versions) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

func Parse(raw string) (SemVer, error) {
	raw = strings.Trim(raw, "\r\n")
	if len(raw) < 6 {
		// e.g. minimal length is 6, e.g. "v1.2.3"
		return zero, fmt.Errorf("%q too short to be a version", raw)
	}
	if raw[0] != 'v' {
		return zero, fmt.Errorf("%q must start with letter 'v'", raw)
	}
	fields := strings.Split(raw[1:], ".")
	if len(fields) < 3 {
		return zero, fmt.Errorf("%q doesn't have the form v1.2.3", raw)
	}
	n := make([]int, 3)
	for i := 0; i < 3; i++ {
		var err error
		n[i], err = strconv.Atoi(fields[i])
		if err != nil {
			return zero, err
		}
	}
	return New(n[0], n[1], n[2]), nil
}

func (v SemVer) Bump(b SvBump) SemVer {
	switch b {
	case Major:
		return New(v.major+1, 0, 0)
	case Minor:
		return New(v.major, v.minor+1, 0)
	case Patch:
		return New(v.major, v.minor, v.patch+1)
	default:
		return New(v.major, v.minor, v.patch+1)
	}
}

func (v SemVer) BranchLabel() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v SemVer) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v SemVer) Pretty() string {
	if v.IsZero() {
		return ""
	}
	return v.String()
}

func (v SemVer) Equals(o SemVer) bool {
	return v.major == o.major && v.minor == o.minor && v.patch == o.patch
}

func (v SemVer) LessThan(o SemVer) bool {
	return v.major < o.major ||
		(v.major == o.major && v.minor < o.minor) ||
		(v.major == o.major && v.minor == o.minor && v.patch < o.patch)
}

func (v SemVer) IsZero() bool {
	return v.Equals(zero)
}
