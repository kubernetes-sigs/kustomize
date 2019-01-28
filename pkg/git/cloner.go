/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/fs"
)

// Cloner is a function that can clone a git repo.
type Cloner func(url string) (*RepoSpec, error)

// ClonerUsingGitExec uses a local git install, as opposed
// to say, some remote API, to obtain a local clone of
// a remote repo.
func ClonerUsingGitExec(spec string) (*RepoSpec, error) {
	gitProgram, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.Wrap(err, "no 'git' program on path")
	}
	repoSpec, err := NewRepoSpecFromUrl(spec)
	if err != nil {
		return nil, err
	}
	repoSpec.cloneDir, err = fs.NewTmpConfirmedDir()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		gitProgram,
		"clone",
		repoSpec.CloneSpec(),
		repoSpec.cloneDir.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "trouble cloning %s", spec)
	}
	if repoSpec.ref == "" {
		return repoSpec, nil
	}
	cmd = exec.Command(gitProgram, "checkout", repoSpec.ref)
	cmd.Dir = repoSpec.cloneDir.String()
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrapf(
			err, "trouble checking out href %s", repoSpec.ref)
	}
	return repoSpec, nil
}
