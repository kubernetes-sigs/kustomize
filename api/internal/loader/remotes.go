// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// remoteCache holds the git clones and OCI pulls of a build, so that a
// remote referenced more than once, under different root paths or from
// different bases, is fetched once. A floating branch or tag then also
// resolves to the same contents for the whole build. The loaders of the
// build share the cache, which removes the clones and pulls on cleanup.
// Like the loaders, it is not safe for concurrent use.
type remoteCache struct {
	dirs     map[string]filesys.ConfirmedDir
	cleaners []func() error
}

func newRemoteCache() *remoteCache {
	return &remoteCache{dirs: make(map[string]filesys.ConfirmedDir)}
}

// gitCacheKey identifies a clone: the repository, the ref, and whether
// the submodules are cloned too.
func gitCacheKey(repoSpec *git.RepoSpec) string {
	return fmt.Sprintf("git %s %s %t", repoSpec.CloneSpec(), repoSpec.Ref, repoSpec.Submodules)
}

// ociCacheKey identifies a pull: the reference of the artifact, by tag or
// by digest.
func ociCacheKey(repoSpec *oci.RepoSpec) string {
	return "oci " + repoSpec.PullSpec()
}

// fetch sets *dir to the directory cached under key, if any. Otherwise it
// runs fetch, which sets *dir, and caches the directory fetched; if fetch
// fails, it runs cleaner. It returns the function that cleans up after the
// caller: none with a cache, which keeps the directory until its own
// cleanup, and cleaner without one (a nil cache), since each caller then
// owns its directory.
func (c *remoteCache) fetch(
	key string, dir *filesys.ConfirmedDir, fetch, cleaner func() error) (func() error, error) {
	if c != nil {
		if cached, found := c.dirs[key]; found {
			*dir = cached
			return noCleanup, nil
		}
	}
	if err := fetch(); err != nil {
		_ = cleaner()
		return nil, err
	}
	if c == nil {
		return cleaner, nil
	}
	c.dirs[key] = *dir
	c.cleaners = append(c.cleaners, cleaner)
	return noCleanup, nil
}

// cleanup removes the clones and pulls cached so far, and returns the
// first error.
func (c *remoteCache) cleanup() error {
	if c == nil {
		return nil
	}
	var first error
	for _, cleaner := range c.cleaners {
		if err := cleaner(); err != nil && first == nil {
			first = err
		}
	}
	c.dirs = make(map[string]filesys.ConfirmedDir)
	c.cleaners = nil
	return first
}

func noCleanup() error { return nil }
