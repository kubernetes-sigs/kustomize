// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	someDigest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	someRepo   = "ghcr.io/some-org/some-repo"
)

func TestNewRepoSpecFromURL(t *testing.T) {
	testcases := map[string]struct {
		input        string
		pullSpec     string
		kustRootPath string
		absPath      string
	}{
		"tag": {
			input:    "oci://ghcr.io/some-org/some-repo:v1.0.0",
			pullSpec: "ghcr.io/some-org/some-repo:v1.0.0",
			absPath:  "/notPulled",
		},
		"tag and root path": {
			input:        "oci://ghcr.io/some-org/some-repo:v1.0.0//path/to/root",
			pullSpec:     "ghcr.io/some-org/some-repo:v1.0.0",
			kustRootPath: "path/to/root",
			absPath:      "/notPulled/path/to/root",
		},
		"tag and single segment root path": {
			input:        "oci://ghcr.io/lazedo/manifests/k8s-dashboard:c89d748//kind",
			pullSpec:     "ghcr.io/lazedo/manifests/k8s-dashboard:c89d748",
			kustRootPath: "kind",
			absPath:      "/notPulled/kind",
		},
		"digest": {
			input:    ociScheme + someRepo + "@" + someDigest,
			pullSpec: someRepo + "@" + someDigest,
			absPath:  "/notPulled",
		},
		"digest and root path": {
			input:        ociScheme + someRepo + "@" + someDigest + "//overlays/dev",
			pullSpec:     someRepo + "@" + someDigest,
			kustRootPath: "overlays/dev",
			absPath:      "/notPulled/overlays/dev",
		},
		"tag and digest": {
			// The digest wins.
			input:    ociScheme + someRepo + ":v1.0.0@" + someDigest,
			pullSpec: someRepo + "@" + someDigest,
			absPath:  "/notPulled",
		},
		"default tag": {
			input:    "oci://ghcr.io/some-org/some-repo",
			pullSpec: "ghcr.io/some-org/some-repo:latest",
			absPath:  "/notPulled",
		},
		"default tag and root path": {
			input:        "oci://ghcr.io/some-org/some-repo//base",
			pullSpec:     "ghcr.io/some-org/some-repo:latest",
			kustRootPath: "base",
			absPath:      "/notPulled/base",
		},
		"empty root path": {
			input:    "oci://ghcr.io/some-org/some-repo:v1.0.0//",
			pullSpec: "ghcr.io/some-org/some-repo:v1.0.0",
			absPath:  "/notPulled",
		},
		"root path with leading slash": {
			input:        "oci://ghcr.io/some-org/some-repo:v1.0.0///base",
			pullSpec:     "ghcr.io/some-org/some-repo:v1.0.0",
			kustRootPath: "base",
			absPath:      "/notPulled/base",
		},
		"root path with trailing slash": {
			input:        "oci://ghcr.io/some-org/some-repo:v1.0.0//base/",
			pullSpec:     "ghcr.io/some-org/some-repo:v1.0.0",
			kustRootPath: "base/",
			absPath:      "/notPulled/base",
		},
		"root path with dot segments staying inside": {
			input:        "oci://ghcr.io/some-org/some-repo:v1.0.0//base/../overlay",
			pullSpec:     "ghcr.io/some-org/some-repo:v1.0.0",
			kustRootPath: "base/../overlay",
			absPath:      "/notPulled/overlay",
		},
		"registry with port": {
			input:        "oci://localhost:5000/some-repo:v1.0.0//base",
			pullSpec:     "localhost:5000/some-repo:v1.0.0",
			kustRootPath: "base",
			absPath:      "/notPulled/base",
		},
		"ip registry with port": {
			input:    "oci://127.0.0.1:5000/some-org/some-repo:v1.0.0",
			pullSpec: "127.0.0.1:5000/some-org/some-repo:v1.0.0",
			absPath:  "/notPulled",
		},
		"docker hub": {
			input:    "oci://docker.io/some-org/some-repo:v1.0.0",
			pullSpec: "index.docker.io/some-org/some-repo:v1.0.0",
			absPath:  "/notPulled",
		},
		"scheme is case insensitive": {
			input:    "OCI://ghcr.io/some-org/some-repo:v1.0.0",
			pullSpec: "ghcr.io/some-org/some-repo:v1.0.0",
			absPath:  "/notPulled",
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			rs, err := NewRepoSpecFromURL(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.input, rs.Raw())
			require.Equal(t, tc.pullSpec, rs.PullSpec())
			require.Equal(t, tc.kustRootPath, rs.KustRootPath)
			require.Equal(t, tc.absPath, rs.AbsPath())
			require.Equal(t, notPulled, rs.PullDir())
		})
	}
}

func TestNewRepoSpecFromURL_Errors(t *testing.T) {
	testcases := map[string]struct {
		input string
		err   string
	}{
		"empty": {
			input: "",
			err:   "uri does not have the oci:// scheme",
		},
		"local path": {
			input: "base",
			err:   "uri does not have the oci:// scheme",
		},
		"absolute path": {
			input: "/some/base",
			err:   "uri does not have the oci:// scheme",
		},
		"git url": {
			input: "https://github.com/some-org/some-repo//base",
			err:   "uri does not have the oci:// scheme",
		},
		"scheme only": {
			input: "oci://",
			err:   "invalid oci artifact reference",
		},
		"no repository": {
			input: "oci://ghcr.io",
			err:   "must include the registry host",
		},
		"no registry": {
			input: "oci://some-org/some-repo:v1.0.0",
			err:   "must include the registry host",
		},
		"root path without delimiter": {
			input: "oci://ghcr.io/some-org/some-repo:v1.0.0/base",
			err:   "the kustomization root path must be separated by //",
		},
		"root path exits artifact": {
			input: "oci://ghcr.io/some-org/some-repo:v1.0.0//../escape",
			err:   "url path exits artifact",
		},
		"root path exits artifact after cleaning": {
			input: "oci://ghcr.io/some-org/some-repo:v1.0.0//base/../../escape",
			err:   "url path exits artifact",
		},
		"bad digest": {
			input: "oci://ghcr.io/some-org/some-repo@sha256:notADigest",
			err:   "invalid oci artifact reference",
		},
		"bad tag": {
			input: "oci://ghcr.io/some-org/some-repo:not a tag",
			err:   "invalid oci artifact reference",
		},
		"query string": {
			input: "oci://ghcr.io/some-org/some-repo?ref=v1.0.0",
			err:   "invalid oci artifact reference",
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, err := NewRepoSpecFromURL(tc.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestRepoSpecCleaner(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	rs, err := NewRepoSpecFromURL("oci://ghcr.io/some-org/some-repo:v1.0.0//base")
	require.NoError(t, err)
	require.NoError(t, DoNothingPuller(filesys.ConfirmedDir("/pulled"))(rs, fSys))
	require.NoError(t, fSys.WriteFile(rs.AbsPath()+"/kustomization.yaml", []byte("whatever")))
	require.True(t, fSys.Exists("/pulled/base/kustomization.yaml"))
	require.NoError(t, rs.Cleaner(fSys)())
	require.False(t, fSys.Exists("/pulled"))
}

func TestIsOciURL(t *testing.T) {
	require.True(t, IsOciURL("oci://ghcr.io/some-org/some-repo:v1.0.0//base"))
	require.True(t, IsOciURL("OCI://ghcr.io/some-org/some-repo"))
	require.True(t, IsOciURL("oci://"))
	require.False(t, IsOciURL("oci:/ghcr.io/some-org/some-repo"))
	require.False(t, IsOciURL("https://github.com/some-org/some-repo//base"))
	require.False(t, IsOciURL("base"))
	require.False(t, IsOciURL(""))
}
