// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/image"
)

func TestIsImageMatched(t *testing.T) {
	testCases := []struct {
		testName  string
		value     string
		name      string
		isMatched bool
	}{
		{
			testName:  "identical",
			value:     "nginx",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is match with tag",
			value:     "nginx:12345",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is match with digest",
			value:     "nginx@sha256:xyz",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is match with tag and digest",
			value:     "nginx:12345@sha256:xyz",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is match with non-sha256 digest",
			value:     "nginx@sha512:xyz",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is match with tag and non-sha256 digest",
			value:     "nginx:12345@sha512:xyz",
			name:      "nginx",
			isMatched: true,
		},
		{
			// Registered SHA-512 algorithm with a full-length digest, as in
			// the OCI image-spec descriptor examples.
			testName:  "name is match with full sha512 digest",
			value:     "nginx@sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			name:      "nginx",
			isMatched: true,
		},
		{
			// Unregistered but OCI-valid algorithm with a separator (+), from
			// the descriptor example "multihash+base58:Qm...".
			testName:  "name is match with multihash+base58 digest",
			value:     "nginx@multihash+base58:QmRZxt2b1FVZPNqd8hsiykDL3TdBDeTSPX9Kv46HmX4Gx8",
			name:      "nginx",
			isMatched: true,
		},
		{
			testName:  "name is not a match",
			value:     "apache:12345",
			name:      "nginx",
			isMatched: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assert.Equal(t, tc.isMatched, image.IsImageMatched(tc.value, tc.name))
		})
	}
}

func TestSplit(t *testing.T) {
	testCases := []struct {
		testName string
		value    string
		name     string
		tag      string
		digest   string
	}{
		{
			testName: "no tag",
			value:    "nginx",
			name:     "nginx",
			tag:      "",
			digest:   "",
		},
		{
			testName: "with tag",
			value:    "nginx:1.2.3",
			name:     "nginx",
			tag:      "1.2.3",
			digest:   "",
		},
		{
			testName: "with digest",
			value:    "nginx@sha256:12345",
			name:     "nginx",
			tag:      "",
			digest:   "sha256:12345",
		},
		{
			testName: "with tag and digest",
			value:    "nginx:1.2.3@sha256:12345",
			name:     "nginx",
			tag:      "1.2.3",
			digest:   "sha256:12345",
		},
		{
			testName: "with domain",
			value:    "docker.io/nginx:1.2.3",
			name:     "docker.io/nginx",
			tag:      "1.2.3",
			digest:   "",
		},
		{
			testName: "with domain and port",
			value:    "foo.com:443/nginx:1.2.3",
			name:     "foo.com:443/nginx",
			tag:      "1.2.3",
			digest:   "",
		},
		{
			testName: "with domain, port, tag and digest",
			value:    "foo.com:443/nginx:1.2.3@sha256:12345",
			name:     "foo.com:443/nginx",
			tag:      "1.2.3",
			digest:   "sha256:12345",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			name, tag, digest := image.Split(tc.value)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.tag, tag)
			assert.Equal(t, tc.digest, digest)
		})
	}
}
