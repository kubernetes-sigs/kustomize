// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"fmt"
	"log"
	"os"
	"strings"
)

// Cloner is a function that can clone a git repo.
type Cloner func(repoSpec *RepoSpec) error

func CachedGitCloner(repoSpec *RepoSpec) error {

	var targetDir filesys.ConfirmedDir
	gitRootDir, err := filesys.GitRootDir()

	if err != nil {
		return err
	}
	gitBranchesRootDir := gitRootDir.Join("branches")

	repoFolderName := strings.ReplaceAll(fmt.Sprintf("%s_%s_%s", repoSpec.Host, repoSpec.OrgRepo, repoSpec.Ref), "/", "_")

	permanentCachedDir := gitRootDir.Join(repoFolderName)
	mutableCachedDir := filesys.ConfirmedDir(gitBranchesRootDir).Join(repoFolderName)

	//search into mutableCachedDir
	if _, err := os.Stat(mutableCachedDir); os.IsNotExist(err) {
		//search into immutable cached dir
		if _, err := os.Stat(permanentCachedDir); os.IsNotExist(err) {
			log.Printf("cloning %s", repoSpec.Raw())
			err := ClonerUsingGitExec(repoSpec)
			if err != nil {
				return err
			}

			// check if it is a "tag" a commitId or simple branch
			isImmutableRef, err := isImmutableRepoRef(repoSpec)
			if err != nil {
				return err
			}

			if !isImmutableRef {
				log.Printf("Git repo %v imported with a branch instead of tag or commitId", repoSpec.raw)
				targetDir = filesys.ConfirmedDir(mutableCachedDir)
			} else {
				targetDir = filesys.ConfirmedDir(permanentCachedDir)
			}

			if _, err := os.Stat(gitRootDir.String()); os.IsNotExist(err) {
				//create .kustomize dir if it does not exist
				err = os.Mkdir(gitRootDir.String(), 0755|os.ModeDir)
				if err != nil {
					return err
				}
			}
			if _, err := os.Stat(gitBranchesRootDir); os.IsNotExist(err) {
				//create .kustomize dir if it does not exist
				err = os.Mkdir(gitBranchesRootDir, 0755|os.ModeDir)
				if err != nil {
					return err
				}
			}
			log.Printf("move %s to %s", repoSpec.Dir.String(), targetDir.String())
			err = os.Rename(repoSpec.Dir.String(), targetDir.String())
			if err != nil {
				return err
			}

		} else { // we found it and it is an immutable
			targetDir = filesys.ConfirmedDir(permanentCachedDir)
		}
	} else {
		//we found it and it is an mutable
		targetDir = filesys.ConfirmedDir(mutableCachedDir)

		//TODO decide if we should clean old copy, update, or let the user clean the cache when he want
	}

	repoSpec.Dir = targetDir
	return nil

}

// check if the git repo ref is not a banch
func isImmutableRepoRef(repoSpec *RepoSpec) (bool, error) {
	r, err := newCmdRunner()
	r.dir = repoSpec.Dir
	if err != nil {
		return false, err
	}
	origin, err := r.run("remote", "show", "origin")
	if err != nil {
		return false, err
	}

	remoteBranches := strings.Split(origin,"Remote branches:")[1]

	return !strings.Contains(remoteBranches, "tracked"), nil
}

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec) error {
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
