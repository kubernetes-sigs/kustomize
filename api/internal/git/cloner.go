// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"os"
	"os/user"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Cloner is a function that can clone a git repo.
type Cloner func(repoSpec *RepoSpec) error

// gitCacheRootDir returns a directory for caching git clones. The directory
// "$XDG_CACHE_HOME/kustomize" is preferred if the environment variable is set,
// otherwise "$HOME/.cache/kustomize" is used. If the user's home dir cannot be
// determined, then "<system temp dir>/kustomize" is used as a fallback.
func gitCacheRootDir() string {
	// Try "$XDG_CACHE_HOME/kustomize".
	if dir := os.Getenv(konfig.XdgCacheHomeEnv); dir != "" {
		return filepath.Join(dir, konfig.ProgramName)
	}

	// Try "$HOME/.cache/kustomize".
	if usr, err := user.Current(); err == nil && usr.HomeDir != "" {
		return filepath.Join(usr.HomeDir, konfig.XdgCacheHomeEnvDefault, konfig.ProgramName)
	}

	// Fallback to "<system temp dir>/kustomize".
	return filepath.Join(os.TempDir(), konfig.ProgramName)
}

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec) error {
	var dir filesys.ConfirmedDir
	var err error
	if repoSpec.Cached {
		// Create a cache directory.
		dir, err = filesys.NewCachedConfirmedDir(gitCacheRootDir(), repoSpec.Host, repoSpec.OrgRepo, repoSpec.Ref)
	} else {
		// Create a temporary directory.
		dir, err = filesys.NewTmpConfirmedDir()
	}

	repoSpec.Dir = dir
	if err == filesys.ErrCachedDirExists {
		// Cached directory already exists, nothing more to do here!
		return nil
	} else if err != nil {
		return err
	}

	r, err := newCmdRunner(dir, repoSpec.Timeout)
	if err != nil {
		return err
	}
	if err = r.run("init"); err != nil {
		return err
	}
	if err = r.run(
		"remote", "add", "origin", repoSpec.CloneSpec()); err != nil {
		return err
	}
	ref := "HEAD"
	if repoSpec.Ref != "" {
		ref = repoSpec.Ref
	}
	if err = r.run("fetch", "--depth=1", "origin", ref); err != nil {
		return err
	}
	if err = r.run("checkout", "FETCH_HEAD"); err != nil {
		return err
	}
	if repoSpec.Submodules {
		return r.run("submodule", "update", "--init", "--recursive")
	}
	return nil
}

// DoNothingCloner returns a cloner that only sets
// cloneDir field in the repoSpec.  It's assumed that
// the cloneDir is associated with some fake filesystem
// used in a test.
func DoNothingCloner(dir filesys.ConfirmedDir) Cloner {
	return func(rs *RepoSpec) error {
		rs.Dir = dir
		return nil
	}
}
