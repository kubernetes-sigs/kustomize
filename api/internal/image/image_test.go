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
