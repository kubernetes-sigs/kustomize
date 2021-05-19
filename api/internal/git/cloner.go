// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"fmt"
	"log"
	"os"
	"os/user"
	"sigs.k8s.io/kustomize/api/konfig"
	"strings"
)

// Cloner is a function that can clone a git repo.
// can allow or not branches references
type Cloner func(repoSpec *RepoSpec, acceptBranches bool) error

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec, acceptBranches bool) error {
	r, err := newCmdRunner(repoSpec.Timeout)
	if err != nil {
		return err
	}
	repoSpec.Dir = r.dir
	if _, err = r.run("init"); err != nil {
		return err
	}
	if _, err = r.run(
		"remote", "add", "origin", repoSpec.CloneSpec()); err != nil {
		return err
	}
	ref := "HEAD"
	if repoSpec.Ref != "" {
		ref = repoSpec.Ref
	}
	if _, err = r.run("fetch", "--depth=1", "origin", ref); err != nil {
		return err
	}
	if _, err = r.run("checkout", "FETCH_HEAD"); err != nil {
		return err
	}
	if repoSpec.Submodules {
		if _, err = r.run("submodule", "update", "--init", "--recursive"); err != nil {
			return err
		}
	}

	// check if it is a "tag", a commitId or simple branch
	repoSpec.IsBranchRef, err = isMutableRepoRef(repoSpec)
	if err != nil {
		return err
	}

	if !acceptBranches && repoSpec.IsBranchRef {
		return fmt.Errorf("we dont accept banch ref %s/%s", repoSpec.OrgRepo, repoSpec.Path)
	}

	return nil
}

// DoNothingCloner returns a cloner that only sets
// cloneDir field in the repoSpec.  It's assumed that
// the cloneDir is associated with some fake filesystem
// used in a test.
func DoNothingCloner(dir filesys.ConfirmedDir) Cloner {
	return func(rs *RepoSpec, _ bool) error {
		rs.Dir = dir
		return nil
	}
}

// CachedGitCloner uses a local cache directory to store cloned refs
// store mutable (branches) ref to "branches" sub-directory to be easy to
// clean later
func CachedGitCloner(repoSpec *RepoSpec, acceptBranches bool) error {

	var targetDir filesys.ConfirmedDir
	gitRootDir, err := gitCacheRootDir()

	if err != nil {
		return err
	}
	gitBranchesRootDir := gitRootDir.Join("branches")

	err = cacheDirPrerequisites(gitRootDir, gitBranchesRootDir)
	if err != nil {
		return err
	}

	repoFolderName := strings.ReplaceAll(fmt.Sprintf("%s_%s_%s", repoSpec.Host, repoSpec.OrgRepo, repoSpec.Ref), "/", "_")

	permanentGitRefPath := gitRootDir.Join(repoFolderName)
	mutableGitRefPath := filesys.ConfirmedDir(gitBranchesRootDir).Join(repoFolderName)

	//search into mutable cached directory
	if _, err := os.Stat(mutableGitRefPath); os.IsNotExist(err) {
		//search into immutable cached dir
		if _, err := os.Stat(permanentGitRefPath); os.IsNotExist(err) {
			targetDir, err = cloneToCacheDir(repoSpec, mutableGitRefPath, permanentGitRefPath, acceptBranches)
			if err != nil {
				return err
			}

		} else { // we found it and it is an immutable
			targetDir = filesys.ConfirmedDir(permanentGitRefPath)
		}
	} else {
		if !acceptBranches {
			return fmt.Errorf("we dont accept banch ref %s/%s", repoSpec.OrgRepo, repoSpec.Path)
		}
		//we found it and it is a mutable
		targetDir = filesys.ConfirmedDir(mutableGitRefPath)

		//TODO decide if we should clean old copy, update, or let the user clean the cache when he want
	}

	repoSpec.Dir = targetDir
	return nil

}

func gitCacheRootDir() (filesys.ConfirmedDir, error) {
	var cacheHome string
	if cacheHome = os.Getenv(konfig.XdgCacheHomeEnv); cacheHome == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		cacheHome = fmt.Sprintf("%s/%s", usr.HomeDir, konfig.XdgCacheHomeEnvDefault)
	}

	return filesys.ConfirmedDir(fmt.Sprintf("%s/%s", cacheHome, "git")), nil
}

func cloneToCacheDir(repoSpec *RepoSpec, mutableGitRefPath, permanentGitRefPath string, acceptBranches bool) (targetDir filesys.ConfirmedDir, err error) {
	log.Printf("cloning %s", repoSpec.Raw())
	err = ClonerUsingGitExec(repoSpec, acceptBranches)
	targetDir = ""
	if err != nil {
		return
	}

	if repoSpec.IsBranchRef {
		log.Printf("Git repo %v imported with a branch instead of tag or commitId", repoSpec.raw)
		targetDir = filesys.ConfirmedDir(mutableGitRefPath)
	} else {
		targetDir = filesys.ConfirmedDir(permanentGitRefPath)
	}

	log.Printf("move %s to %s", repoSpec.Dir.String(), targetDir.String())
	err = os.Rename(repoSpec.Dir.String(), targetDir.String())
	if err != nil {
		return
	}
	return
}

func cacheDirPrerequisites(gitRootDir filesys.ConfirmedDir, gitBranchesRootDir string) error {
	if _, err := os.Stat(gitRootDir.String()); os.IsNotExist(err) {
		//create permanent cache dir if it does not exist
		err = os.Mkdir(gitRootDir.String(), 0755|os.ModeDir)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(gitBranchesRootDir); os.IsNotExist(err) {
		//create mutable ref cache dir if it does not exist
		err = os.Mkdir(gitBranchesRootDir, 0755|os.ModeDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// check if the git repo ref is a branch
func isMutableRepoRef(repoSpec *RepoSpec) (bool, error) {
	r, err := newCmdRunner(repoSpec.Timeout)
	r.dir = repoSpec.Dir
	if err != nil {
		return false, err
	}
	origin, err := r.run("remote", "show", "origin")
	if err != nil {
		return false, err
	}

	remoteBranches := strings.Split(origin, "Remote branches:")[1]

	return strings.Contains(remoteBranches, "tracked"), nil
}
