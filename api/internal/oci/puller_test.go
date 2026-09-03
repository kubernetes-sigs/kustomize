// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci_test

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/api/internal/oci/ocitest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	ociScheme = "oci://"
	someRepo  = "/some-org/some-repo:v1.0.0"
)

func artifactFiles() map[string]string {
	return map[string]string{
		"kustomization.yaml":      "resources:\n- deployment.yaml\n",
		"deployment.yaml":         "kind: Deployment\n",
		"overlays/dev/patch.yaml": "kind: Deployment\n",
	}
}

func pull(t *testing.T, url string) (*oci.RepoSpec, filesys.FileSystem) {
	t.Helper()
	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL(url)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, rs.Cleaner(fSys)()) })
	require.NoError(t, oci.PullerUsingRegistry(rs, fSys))
	return rs, fSys
}

func requireFiles(t *testing.T, fSys filesys.FileSystem, dir filesys.ConfirmedDir, files map[string]string) {
	t.Helper()
	for path, content := range files {
		data, err := fSys.ReadFile(dir.Join(path))
		require.NoError(t, err, path)
		require.Equal(t, content, string(data), path)
	}
}

func TestPullerUsingRegistry(t *testing.T) {
	host := ocitest.StartRegistry(t)
	testcases := map[string]ocitest.Artifact{
		"oci uncompressed layer": ocitest.Files(artifactFiles()),
		"oci gzip layer": {
			Entries:  ocitest.Files(artifactFiles()).Entries,
			Compress: true,
		},
		"flux artifact": ocitest.FluxArtifact(artifactFiles()),
	}
	for name, artifact := range testcases {
		t.Run(name, func(t *testing.T) {
			ref := host + someRepo
			digest := ocitest.Push(t, ref, artifact)

			t.Run("by tag", func(t *testing.T) {
				rs, fSys := pull(t, ociScheme+ref+"//overlays/dev")
				require.NotEqual(t, "/notPulled", rs.PullDir().String())
				require.Equal(t, rs.PullDir().Join("overlays/dev"), rs.AbsPath())
				requireFiles(t, fSys, rs.PullDir(), artifactFiles())
			})
			t.Run("by digest", func(t *testing.T) {
				rs, fSys := pull(t, ociScheme+host+"/some-org/some-repo@"+digest)
				requireFiles(t, fSys, rs.PullDir(), artifactFiles())
			})
		})
	}
}

func TestPullerUsingRegistry_CleanerRemovesPullDir(t *testing.T) {
	host := ocitest.StartRegistry(t)
	ref := host + someRepo
	ocitest.Push(t, ref, ocitest.Files(artifactFiles()))

	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL(ociScheme + ref)
	require.NoError(t, err)
	require.NoError(t, oci.PullerUsingRegistry(rs, fSys))
	require.True(t, fSys.Exists(rs.PullDir().Join("kustomization.yaml")))
	require.NoError(t, rs.Cleaner(fSys)())
	require.False(t, fSys.Exists(rs.PullDir().String()))
}

func TestPullerUsingRegistry_MissingArtifact(t *testing.T) {
	host := ocitest.StartRegistry(t)
	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL(ociScheme + host + "/some-org/missing:v1.0.0")
	require.NoError(t, err)
	err = oci.PullerUsingRegistry(rs, fSys)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pulling oci artifact "+host+"/some-org/missing:v1.0.0")
	// The pull dir is created before the pull and must be cleaned by the caller.
	require.True(t, fSys.Exists(rs.PullDir().String()))
	require.NoError(t, rs.Cleaner(fSys)())
	require.False(t, fSys.Exists(rs.PullDir().String()))
}

func TestPullerUsingRegistry_SkipsLinks(t *testing.T) {
	host := ocitest.StartRegistry(t)
	ref := host + someRepo
	artifact := ocitest.Files(artifactFiles())
	artifact.Entries["escape"] = ocitest.Entry{Typeflag: tar.TypeSymlink, Linkname: "../../../etc"}
	artifact.Entries["link"] = ocitest.Entry{Typeflag: tar.TypeLink, Linkname: "deployment.yaml"}
	artifact.Entries["empty/"] = ocitest.Entry{Typeflag: tar.TypeDir}
	ocitest.Push(t, ref, artifact)

	rs, fSys := pull(t, ociScheme+ref)
	requireFiles(t, fSys, rs.PullDir(), artifactFiles())
	require.False(t, fSys.Exists(rs.PullDir().Join("escape")))
	require.False(t, fSys.Exists(rs.PullDir().Join("link")))
	require.True(t, fSys.IsDir(rs.PullDir().Join("empty")))
}

func TestPullerUsingRegistry_RejectsNonLocalEntries(t *testing.T) {
	host := ocitest.StartRegistry(t)
	for name, tc := range map[string]struct {
		path string
		errs []string // any of these
	}{
		// Rejected by go-containerregistry while flattening the layers
		// (newer releases) or by the puller's own check (older ones).
		"parent": {"../escape.yaml", []string{
			"reading contents of oci artifact", "is not local to the artifact"}},
		// Passed through by go-containerregistry, rejected here.
		"absolute": {"/escape.yaml", []string{"is not local to the artifact"}},
	} {
		t.Run(name, func(t *testing.T) {
			ref := host + "/some-org/" + name + ":v1.0.0"
			ocitest.Push(t, ref, ocitest.Files(map[string]string{tc.path: "kind: Deployment\n"}))

			fSys := filesys.MakeFsOnDisk()
			rs, err := oci.NewRepoSpecFromURL(ociScheme + ref)
			require.NoError(t, err)
			err = oci.PullerUsingRegistry(rs, fSys)
			require.Error(t, err)
			require.Condition(t, func() bool {
				for _, e := range tc.errs {
					if strings.Contains(err.Error(), e) {
						return true
					}
				}
				return false
			}, "error %q matches none of %q", err.Error(), tc.errs)
			require.Contains(t, err.Error(), "escape.yaml")
			require.NoError(t, rs.Cleaner(fSys)())
			_, err = os.Stat(filepath.Join(filepath.Dir(rs.PullDir().String()), "escape.yaml"))
			require.True(t, os.IsNotExist(err))
		})
	}
}
