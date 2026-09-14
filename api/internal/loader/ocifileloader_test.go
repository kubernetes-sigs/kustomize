// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	ociTopDir   = "/ociTop"
	ociPullRoot = "/ociPull"
	somePull    = "/somePull"
	baseDir     = "/base"
	overlayDir  = "/overlay"
)

func TestNewLoaderAtOciPull(t *testing.T) {
	require := require.New(t)

	rootURL := "oci://ghcr.io/some-org/some-repo:v1.0.0"
	pathInRepo := "foo/base"
	url := rootURL + "//" + pathInRepo
	pullRoot := ociPullRoot
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(pullRoot))
	require.NoError(fSys.MkdirAll(pullRoot + "/" + pathInRepo))
	require.NoError(fSys.WriteFile(
		pullRoot+"/"+pathInRepo+"/"+
			konfig.DefaultKustomizationFileName(),
		[]byte(`
whatever
`)))

	repoSpec, err := oci.NewRepoSpecFromURL(url)
	require.NoError(err)

	l, err := newLoaderAtOciPull(
		repoSpec, fSys, nil,
		oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))
	require.NoError(err)
	repo := l.Repo()
	require.Equal(pullRoot, repo)
	require.Equal(pullRoot+"/"+pathInRepo, l.Root())

	// Same artifact: a cycle.
	_, err = l.New(url)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")

	// Relative to an artifact, but not inside it.
	_, err = l.New("../../..")
	require.Error(err)

	pathInRepo = "foo/overlay"
	require.NoError(fSys.MkdirAll(pullRoot + "/" + pathInRepo))
	url = "oci://ghcr.io/some-org/other-repo:v1.0.0//" + pathInRepo
	l2, err := l.New(url)
	require.NoError(err)

	repo = l2.Repo()
	require.Equal(pullRoot, repo)
	require.Equal(pullRoot+"/"+pathInRepo, l2.Root())

	// Relative paths inside the artifact are fine.
	l3, err := l.New("../overlay")
	require.NoError(err)
	require.Equal(pullRoot+"/"+pathInRepo, l3.Root())
}

func TestNewLoaderAtOciPullFileNotDir(t *testing.T) {
	require := require.New(t)

	pullRoot := ociPullRoot
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(pullRoot))
	require.NoError(fSys.WriteFile(pullRoot+"/deployment.yaml", []byte("kind: Deployment\n")))

	repoSpec, err := oci.NewRepoSpecFromURL("oci://ghcr.io/some-org/some-repo:v1.0.0//deployment.yaml")
	require.NoError(err)
	_, err = newLoaderAtOciPull(
		repoSpec, fSys, nil,
		oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))
	require.Error(err)
	require.Contains(err.Error(), "expecting directory")
}

func TestNewLoaderAtOciPullFails(t *testing.T) {
	require := require.New(t)

	fSys := filesys.MakeFsInMemory()
	pullRoot := ociPullRoot
	require.NoError(fSys.MkdirAll(pullRoot))
	pullErr := fmt.Errorf("no such artifact")
	puller := func(rs *oci.RepoSpec, _ filesys.FileSystem) error {
		rs.Dir = filesys.ConfirmedDir(pullRoot)
		return pullErr
	}

	repoSpec, err := oci.NewRepoSpecFromURL("oci://ghcr.io/some-org/some-repo:v1.0.0")
	require.NoError(err)
	_, err = newLoaderAtOciPull(repoSpec, fSys, nil, puller)
	require.ErrorIs(err, pullErr)
	// The pull dir is cleaned up on failure.
	require.False(fSys.Exists(pullRoot))
}

func TestLoaderDisallowsLocalBaseFromOciOverlay(t *testing.T) {
	// Define an overlay-base structure in the file system,
	// pull the overlay from an artifact, and
	// then confirm that overlay cannot load base.
	require := require.New(t)

	topDir := ociTopDir
	pullRoot := topDir + somePull
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(topDir + baseDir))
	require.NoError(fSys.MkdirAll(pullRoot + overlayDir))

	repoSpec, err := oci.NewRepoSpecFromURL("oci://ghcr.io/some-org/some-repo:v1.0.0//overlay")
	require.NoError(err)
	l1, err := newLoaderAtOciPull(
		repoSpec, fSys, nil,
		oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot)))
	require.NoError(err)
	require.Equal(pullRoot+overlayDir, l1.Root())

	_, err = l1.New("../../base")
	require.Error(err)
	require.Contains(err.Error(), "must be within the artifact")
}

func TestLocalLoaderReferencingOciBase(t *testing.T) {
	require := require.New(t)

	topDir := ociTopDir
	pullRoot := topDir + somePull
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(topDir))
	require.NoError(fSys.MkdirAll(pullRoot + "/foo/base"))

	l1 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil, nil)
	l1.puller = oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot))
	require.Equal(topDir, l1.Root())

	l2, err := l1.New("oci://ghcr.io/some-org/some-repo:v1.0.0//foo/base")
	require.NoError(err)
	repo := l2.Repo()
	require.Equal(pullRoot, repo)
	require.Equal(pullRoot+"/foo/base", l2.Root())
}

func TestOciRepoIndirectCycleDetection(t *testing.T) {
	require := require.New(t)

	topDir := ociTopDir
	pullRoot := topDir + somePull
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(topDir))
	require.NoError(fSys.MkdirAll(pullRoot))

	l0 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil, nil)
	l0.puller = oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot))

	p1 := "oci://ghcr.io/some-org/some-repo1:v1.0.0"
	p2 := "oci://ghcr.io/some-org/some-repo2:v1.0.0"

	l1, err := l0.New(p1)
	require.NoError(err)

	l2, err := l1.New(p2)
	require.NoError(err)

	_, err = l2.New(p1)
	require.Error(err)
	require.Contains(err.Error(), "cycle detected")
}

func TestNewLoaderOciTakesPrecedenceOverGit(t *testing.T) {
	// The git parser must not claim oci URLs, and vice versa.
	_, err := git.NewRepoSpecFromURL("oci://ghcr.io/some-org/some-repo@sha256:" + strings.Repeat("1", 64) + "//base")
	require.Error(t, err)
	_, err = oci.NewRepoSpecFromURL("https://github.com/some-org/some-repo//base?ref=v1.0.0")
	require.Error(t, err)
}

func TestLoaderRelativeBaseWithinNestedRemotes(t *testing.T) {
	// A git clone refers to an OCI artifact, which refers to a relative
	// base, and vice versa: containment is checked against the nearest
	// remote root only.
	require := require.New(t)

	topDir := ociTopDir
	cloneRoot := topDir + "/someClone"
	pullRoot := topDir + somePull
	fSys := filesys.MakeFsInMemory()
	require.NoError(fSys.MkdirAll(cloneRoot + overlayDir))
	require.NoError(fSys.MkdirAll(cloneRoot + baseDir))
	require.NoError(fSys.MkdirAll(pullRoot + overlayDir))
	require.NoError(fSys.MkdirAll(pullRoot + baseDir))
	l0 := newLoaderAtConfirmedDir(
		RestrictionRootOnly, filesys.ConfirmedDir(topDir), fSys, nil,
		git.DoNothingCloner(filesys.ConfirmedDir(cloneRoot)))
	l0.puller = oci.DoNothingPuller(filesys.ConfirmedDir(pullRoot))

	gitOverlay, err := l0.New("github.com/some-org/some-repo/overlay")
	require.NoError(err)
	ociOverlay, err := gitOverlay.New("oci://ghcr.io/some-org/some-repo:v1.0.0//overlay")
	require.NoError(err)
	ociBase, err := ociOverlay.New("../base")
	require.NoError(err)
	require.Equal(pullRoot+baseDir, ociBase.Root())
	_, err = ociOverlay.New("../../someClone/base")
	require.Error(err)
	require.Contains(err.Error(), "must be within the artifact")

	ociOverlay, err = l0.New("oci://ghcr.io/some-org/some-repo:v1.0.0//overlay")
	require.NoError(err)
	gitOverlay, err = ociOverlay.New("github.com/some-org/some-repo/overlay")
	require.NoError(err)
	gitBase, err := gitOverlay.New("../base")
	require.NoError(err)
	require.Equal(cloneRoot+baseDir, gitBase.Root())
	_, err = gitOverlay.New("../../somePull/base")
	require.Error(err)
	require.Contains(err.Error(), "must be within the repo")
}
