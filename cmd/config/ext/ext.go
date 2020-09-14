// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package ext

// KRMFileName returns the name of the KRM file. KRM file determines package
// boundaries and contains the openapi information for a package.
var KRMFileName = func() string {
	return "Krmfile"
}
