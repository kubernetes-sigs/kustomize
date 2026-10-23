// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"regexp"
	"strings"
)

// IsImageMatched returns true if the value of t is identical to the
// image name in the full image name and tag as given by s.
func IsImageMatched(s, t string) bool {
	// Tag values are limited to [a-zA-Z0-9_.{}-].
	// Some tools like Bazel rules_k8s allow tag patterns with {} characters.
	// More info: https://github.com/bazelbuild/rules_k8s/pull/423
	//
	// The digest algorithm is matched following the OCI grammar
	// (algorithm-component separated by one of [+._-], e.g. sha256, sha512,
	// multihash+base58) rather than hard-coded to sha256, so that references
	// using any OCI-valid digest algorithm match consistently with Split,
	// which accepts any algorithm.
	// See https://github.com/opencontainers/image-spec/blob/main/descriptor.md#digests
	// The name t comes from kustomization images[].name and is interpolated
	// into the pattern directly, so it can be an invalid regexp (for example
	// "["). When it fails to compile, treat it as matching nothing rather than
	// dereferencing a nil *Regexp, which would panic during the build.
	pattern, err := regexp.Compile("^" + t + "(:[a-zA-Z0-9_.{}-]*)?(@[a-zA-Z0-9]+([.+_-][a-zA-Z0-9]+)*:[a-zA-Z0-9_.{}-]*)?$")
	if err != nil {
		return false
	}
	return pattern.MatchString(s)
}

// Split separates and returns the name, tag and digest parts
// from the image string using either colon `:` or at `@` separators.
func Split(imageName string) (name string, tag string, digest string) {
	// First, handle digest by splitting at '@'
	parts := strings.SplitN(imageName, "@", 2)
	base := parts[0]
	if len(parts) > 1 {
		digest = parts[1]
	}

	// Find the last slash and last colon in the base string
	lastSlash := strings.LastIndex(base, "/")
	lastColon := strings.LastIndex(base, ":")

	// Ensure the colon is part of the image tag (i.e., after the last slash, avoiding port colons)
	if lastColon > lastSlash {
		tag = base[lastColon+1:]
		name = base[:lastColon]
	} else {
		name = base
	}

	return name, tag, digest
}