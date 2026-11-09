// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// fakeRemotes clones and pulls into a new directory of the file system
// each time, /remote0, /remote1 and so on, holding the given directories,
// and records the URLs it fetched.
type fakeRemotes struct {
	fSys    filesys.FileSystem
	dirs    []string
	fetched []string
}

func (f *fakeRemotes) fetch(url string, dir *filesys.ConfirmedDir) error {
	*dir = filesys.ConfirmedDir(fmt.Sprintf("/remote%d", len(f.fetched)))
	f.fetched = append(f.fetched, url)
	for _, sub := range f.dirs {
		if err := f.fSys.MkdirAll(dir.Join(sub)); err != nil {
			return fmt.Errorf("fake fetch: %w", err)
		}
	}
	return nil
}

func (f *fakeRemotes) cloner(rs *git.RepoSpec) error {
	return f.fetch(rs.Raw(), &rs.Dir)
}

func (f *fakeRemotes) puller(rs *oci.RepoSpec, _ filesys.FileSystem, _ *http.Client) error {
	return f.fetch(rs.Raw(), &rs.Dir)
}

// newTopLoader returns a loader at /top that fetches remotes with fake,
// sharing them through the given cache, if any.
func newTopLoader(t *testing.T, fake *fakeRemotes, remotes *remoteCache) *FileLoader {
	t.Helper()
	require.NoError(t, fake.fSys.MkdirAll("/top"))
	return newLoaderAtConfirmedDir(RestrictionRootOnly, "/top", fake.fSys, nil,
		fake.cloner, fake.puller, remotes)
}

func mustNew(t *testing.T, ldr ifc.Loader, path string) ifc.Loader {
	t.Helper()
	child, err := ldr.New(path)
	require.NoError(t, err)
	return child
}

func TestRemotesFetchedOncePerBuild(t *testing.T) {
	require := require.New(t)
	fake := &fakeRemotes{fSys: filesys.MakeFsInMemory(), dirs: []string{"a", "b"}}
	top := newTopLoader(t, fake, newRemoteCache())

	gitA := mustNew(t, top, "github.com/org/repo/a")
	gitB := mustNew(t, top, "github.com/org/repo/b")
	ociA := mustNew(t, top, "oci://ghcr.io/org/repo:v1//a")
	ociB := mustNew(t, gitA, "oci://ghcr.io/org/repo:v1//b")
	others := []ifc.Loader{
		mustNew(t, top, "github.com/org/repo/a?ref=v2"),
		mustNew(t, top, "github.com/org/repo/a?submodules=false"),
		mustNew(t, top, "oci://ghcr.io/org/repo:v2//a"),
	}

	// Other root paths of the same repository ref or artifact, even from
	// another base, share its clone or pull. Other refs, tags or submodule
	// settings get their own.
	require.Equal("/remote0", gitB.Repo())
	require.Equal("/remote0/b", gitB.Root())
	require.Equal("/remote1", ociB.Repo())
	require.Equal("/remote1/b", ociB.Root())
	require.Equal([]string{
		"github.com/org/repo/a",
		"oci://ghcr.io/org/repo:v1//a",
		"github.com/org/repo/a?ref=v2",
		"github.com/org/repo/a?submodules=false",
		"oci://ghcr.io/org/repo:v2//a",
	}, fake.fetched)

	// The loaders below the top one leave the clones and pulls in place,
	// and cleaning up the top one removes them all.
	for _, ldr := range append(others, gitA, gitB, ociA, ociB) {
		require.NoError(ldr.Cleanup())
	}
	for i := range fake.fetched {
		require.True(fake.fSys.Exists(fmt.Sprintf("/remote%d", i)))
	}
	require.NoError(top.Cleanup())
	for i := range fake.fetched {
		require.False(fake.fSys.Exists(fmt.Sprintf("/remote%d", i)))
	}
}

func TestRemotesTargetInCache(t *testing.T) {
	// A remote build target shares its clone with the references to the
	// same repository ref, and is removed with them.
	require := require.New(t)
	fake := &fakeRemotes{fSys: filesys.MakeFsInMemory(), dirs: []string{"a", "b"}}
	remotes := newRemoteCache()
	repoSpec, err := git.NewRepoSpecFromURL("github.com/org/repo/a")
	require.NoError(err)
	top, err := newLoaderAtGitClone(repoSpec, fake.fSys, nil, fake.cloner, fake.puller, remotes)
	require.NoError(err)

	b := mustNew(t, top, "github.com/org/repo/b")
	require.Equal(top.Repo(), b.Repo())
	require.Len(fake.fetched, 1)
	require.NoError(b.Cleanup())
	require.True(fake.fSys.Exists("/remote0"))
	require.NoError(top.Cleanup())
	require.False(fake.fSys.Exists("/remote0"))
}

func TestRemotesWithoutCache(t *testing.T) {
	// Without a cache, as below NewLoaderOrDie, every reference is fetched,
	// and its loader removes the clone or pull on cleanup.
	require := require.New(t)
	fake := &fakeRemotes{fSys: filesys.MakeFsInMemory(), dirs: []string{"a", "b"}}
	top := newTopLoader(t, fake, nil)

	gitA := mustNew(t, top, "github.com/org/repo/a")
	gitB := mustNew(t, top, "github.com/org/repo/b")
	ociA := mustNew(t, top, "oci://ghcr.io/org/repo:v1//a")
	ociB := mustNew(t, top, "oci://ghcr.io/org/repo:v1//b")
	require.Len(fake.fetched, 4)
	for _, ldr := range []ifc.Loader{gitA, gitB, ociA, ociB} {
		require.True(fake.fSys.Exists(ldr.Repo()))
		require.NoError(ldr.Cleanup())
		require.False(fake.fSys.Exists(ldr.Repo()))
	}
	require.NoError(top.Cleanup())
}

func TestRemotesFailedFetchNotCached(t *testing.T) {
	// A failed clone is removed, and the next reference clones again.
	require := require.New(t)
	fake := &fakeRemotes{fSys: filesys.MakeFsInMemory(), dirs: []string{"a"}}
	failures := 1
	cloner := func(rs *git.RepoSpec) error {
		if failures == 0 {
			return fake.cloner(rs)
		}
		failures--
		rs.Dir = "/failed"
		require.NoError(fake.fSys.MkdirAll("/failed"))
		return fmt.Errorf("network down")
	}
	require.NoError(fake.fSys.MkdirAll("/top"))
	top := newLoaderAtConfirmedDir(RestrictionRootOnly, "/top", fake.fSys, nil,
		cloner, fake.puller, newRemoteCache())

	_, err := top.New("github.com/org/repo/a")
	require.ErrorContains(err, "network down")
	require.False(fake.fSys.Exists("/failed"))
	mustNew(t, top, "github.com/org/repo/a")
	mustNew(t, top, "github.com/org/repo/a")
	require.Len(fake.fetched, 1)
	require.NoError(top.Cleanup())
}

func TestRemotesSharedCloneCycles(t *testing.T) {
	// An overlay of a repository refers to a base of the same repository
	// ref, which refers to the directory above the overlay: as when each
	// reference had a clone of its own, that is no cycle. Cycles within a
	// reference are still detected.
	require := require.New(t)
	fake := &fakeRemotes{fSys: filesys.MakeFsInMemory(), dirs: []string{"deep/overlay", "base"}}
	top := newTopLoader(t, fake, newRemoteCache())

	overlay := mustNew(t, top, "github.com/org/repo/deep/overlay")
	base := mustNew(t, overlay, "github.com/org/repo/base")
	require.Equal(overlay.Repo(), base.Repo())
	deep := mustNew(t, base, "../deep")
	require.Equal("/remote0/deep", deep.Root())

	_, err := deep.New("../base")
	require.ErrorContains(err, "cycle detected")
	_, err = mustNew(t, overlay, "../../base").New("../deep/overlay")
	require.ErrorContains(err, "cycle detected")
	require.NoError(top.Cleanup())
}
