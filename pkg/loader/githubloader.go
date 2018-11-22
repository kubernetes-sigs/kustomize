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

package loader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter"
)

// gitCloner is a function that can clone a git repo.
type gitCloner func(url string) (
	// Directory where the repo is cloned to.
	checkoutDir string,
	// Relative path in the checkoutDir to location
	// of kustomization file.
	pathInCoDir string,
	// Any error encountered when cloning.
	err error)

func makeTmpDir() (string, error) {
	return ioutil.TempDir("", "kustomize-")
}

func hashicorpGitCloner(repoUrl string) (
	checkoutDir string, pathInCoDir string, err error) {
	dir, err := makeTmpDir()
	if err != nil {
		return
	}
	checkoutDir = filepath.Join(dir, "repo")
	url, pathInCoDir := getter.SourceDirSubdir(repoUrl)
	err = checkout(url, checkoutDir)
	return
}

// isRepoUrl checks if a string is a repo Url
func isRepoUrl(s string) bool {
	if strings.HasPrefix(s, "https://") {
		return true
	}
	if strings.HasPrefix(s, "git::") {
		return true
	}
	host := strings.SplitN(s, "/", 2)[0]
	return strings.Contains(host, ".com") || strings.Contains(host, ".org")
}

// Checkout clones a github repo with specified commit/tag/branch
func checkout(url, dir string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	client := &getter.Client{
		Src:  url,
		Dst:  dir,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}
	return client.Get()
}
