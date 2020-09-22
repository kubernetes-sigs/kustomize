// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"log"
	"os/exec"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filesys"
)

// Cloner is a function that can clone a git repo.
type Cloner func(repoSpec *RepoSpec) error

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(repoSpec *RepoSpec) error {
	gitProgram, err := exec.LookPath("git")
	if err != nil {
		return errors.Wrap(err, "no 'git' program on path")
	}
	repoSpec.Dir, err = filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		gitProgram,
		"clone",
		"--depth=1",
		repoSpec.CloneSpec(),
		repoSpec.Dir.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error cloning git repo: %s", out)
		return errors.Wrapf(
			err,
			"trouble cloning git repo %v in %s",
			repoSpec.CloneSpec(), repoSpec.Dir.String())
	}

	if repoSpec.Ref != "" {
		cmd = exec.Command(
			gitProgram,
			"fetch",
			"--depth=1",
			"origin",
			repoSpec.Ref)
		cmd.Dir = repoSpec.Dir.String()
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error fetching ref: %s", out)
			return errors.Wrapf(err, "trouble fetching %s", repoSpec.Ref)
		}

		cmd = exec.Command(
			gitProgram,
			"checkout",
			"FETCH_HEAD")
		cmd.Dir = repoSpec.Dir.String()
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error checking out ref: %s", out)
			return errors.Wrapf(err, "trouble checking out %s", repoSpec.Ref)
		}
	}

	cmd = exec.Command(
		gitProgram,
		"submodule",
		"update",
		"--init",
		"--recursive")
	cmd.Dir = repoSpec.Dir.String()
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error fetching submodules: %s", out)
		return errors.Wrapf(err, "trouble fetching submodules for %s", repoSpec.CloneSpec())
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
