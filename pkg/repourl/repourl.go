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

// Package repourl provides functions to parse a remote repo url
package repourl

import (
	"fmt"
	"regexp"
)

var recognizedRepos = []*regexp.Regexp{
	regexp.MustCompile("^(git@|)([^:@]+):([^/]+)/([^.#/ ]+)(.git|)(|/([0-9A-Za-z-_/]+))(|#([a-zA-Z0-9._-]*))$"),
	regexp.MustCompile("^(https://|)([^/@]+)/([^/]+)/([^.#/ ]+)(.git|)(|/([0-9A-Za-z-_/]+))(|#([a-zA-Z0-9._-]*))$"),
}

// Repo defines a remote repo target
// url must be specified
// ref is optional, could be commit, tag or branch
type Repo struct {
	url string
	ref string
	dir string
}

// NewRepo create a Repo object from a repoUrl
/*
   The accepted repoUrl has following format
   (git@|https://|)github.com/<project>/<repository>(.git|)(#ref)

   The following repoUrls are all accepted:
   git@github.com:kubernetes-sigs/kustomize.git
   https://github.com/kubernetes-sigs/kustomize.git
   https://github.com/kubernetes-sigs/kustomize
   https://github.com/kubernetes-sigs/kustomize#v1.0.6
   github.com/kubernetes-sigs/kustomize#017c4ae0aa19195db2a51ecc5aa82c56a1f1c99b
   git@github.com:kubernetes-sigs/kustomize.git#test-branch
*/
func NewRepo(repoUrl string) (*Repo, error) {
	for _, r := range recognizedRepos {
		matches := r.FindStringSubmatch(repoUrl)
		if len(matches) == 0 {
			continue
		}
		repo := fmt.Sprintf("https://%s/%s/%s.git", matches[2], matches[3], matches[4])
		return &Repo{url: repo, ref: matches[9], dir: matches[7]}, nil
	}
	return nil, fmt.Errorf("%s is not valid github repo", repoUrl)
}
