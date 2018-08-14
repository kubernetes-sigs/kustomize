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

package repourl

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Checkout clones a github repo and checkout the specified commit/tag/branch
func Checkout(url, dir string) (string, error) {
	repo, err := NewRepo(url)
	if err != nil {
		return dir, err
	}

	cmd := exec.Command("git", "clone", repo.url, dir)
	err = cmd.Run()
	if err != nil {
		return dir, err
	}

	log.Printf("Checked out %s into %s", repo.url, dir)

	if repo.ref == "" {
		if repo.dir != "" {
			return filepath.Join(dir, repo.dir), nil
		}
		return dir, nil
	}

	cmd = exec.Command("git", "checkout", repo.ref)
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return dir, err
	}

	if repo.dir != "" {
		return filepath.Join(dir, repo.dir), nil
	}
	return dir, nil
}

// IsRepoUrl checks if a string is a repo Url
func IsRepoUrl(s string) bool {
	return strings.Contains(s, "github.com") || strings.Contains(s, "https://") || strings.Contains(s, "git@")
}
