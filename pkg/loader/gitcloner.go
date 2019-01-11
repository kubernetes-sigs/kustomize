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
	"bytes"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
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

// isRepoUrl checks if a string is likely a github repo Url.
func isRepoUrl(arg string) bool {
	arg = strings.ToLower(arg)
	return !filepath.IsAbs(arg) &&
		(strings.HasPrefix(arg, "git::") ||
			strings.HasPrefix(arg, "gh:") ||
			strings.HasPrefix(arg, "github.com") ||
			strings.HasPrefix(arg, "git@") ||
			strings.Index(arg, "github.com/") > -1 ||
			isAzureHost(arg) || isAWSHost(arg))
}

func makeTmpDir() (string, error) {
	return ioutil.TempDir("", "kustomize-")
}

func simpleGitCloner(spec string) (
	checkoutDir string, pathInCoDir string, err error) {
	gitProgram, err := exec.LookPath("git")
	if err != nil {
		return "", "", errors.Wrap(err, "no 'git' program on path")
	}
	checkoutDir, err = makeTmpDir()
	if err != nil {
		return
	}
	repo, pathInCoDir, gitRef, err := parseUrl(spec)
	if err != nil {
		return
	}
	cmd := exec.Command(
		gitProgram,
		"clone",
		repo,
		checkoutDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", "",
			errors.Wrapf(err, "trouble cloning %s", spec)
	}
	if gitRef == "" {
		return
	}
	cmd = exec.Command(gitProgram, "checkout", gitRef)
	cmd.Dir = checkoutDir
	err = cmd.Run()
	if err != nil {
		return "", "",
			errors.Wrapf(err, "trouble checking out href %s", gitRef)
	}
	return checkoutDir, pathInCoDir, nil
}

func parseUrl(n string) (
	repo string, path string, gitRef string, err error) {
	host, repo, path, gitRef, err := parseGithubUrl(n)
	if err != nil {
		return
	}
	if isAzureHost(host) || isAWSHost(host) {
		repo = host + repo
		return
	}
	repo = host + repo + ".git"
	return
}

const (
	refQuery  = "?ref="
	gitSuffix = ".git"
)

// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the parts.
func parseGithubUrl(n string) (
	host string, repo string, path string, gitRef string, err error) {
	host, n = parseHostSpec(n)
	host = normalizeGitHostSpec(host)

	if strings.HasSuffix(n, gitSuffix) {
		repo = n[0 : len(n)-len(gitSuffix)]
		return
	}
	if strings.Contains(n, gitSuffix) {
		index := strings.Index(n, gitSuffix)
		repo = n[0:index]
		n = n[index+len(gitSuffix):]
		path, gitRef = peelQuery(n)
		return
	}
	i := strings.Index(n, "/")
	if i < 1 {
		return "", "", "", "", errors.New("no separator")
	}
	j := strings.Index(n[i+1:], "/")
	if j >= 0 {
		j += i + 1
		repo = n[:j]
		path, gitRef = peelQuery(n[j+1:])
	} else {
		path = ""
		repo, gitRef = peelQuery(n)
	}
	return
}

func peelQuery(arg string) (string, string) {
	j := strings.Index(arg, refQuery)
	if j >= 0 {
		return arg[:j], arg[j+len(refQuery):]
	}
	return arg, ""
}

func parseHostSpec(n string) (string, string) {
	var host string
	for _, p := range []string{
		// Order matters here.
		"git::", "gh:", "ssh://", "https://", "http://",
		"git@", "github.com:", "github.com/", "gitlab.com/"} {
		if strings.ToLower(n[:len(p)]) == p {
			n = n[len(p):]
			host = host + p
		}
	}
	for _, p := range []string{
		"git-codecommit.[a-z0-9-]*.amazonaws.com/",
		"dev.azure.com/",
		".*visualstudio.com/"} {
		index := regexp.MustCompile(p).FindStringIndex(n)
		if len(index) > 0 {
			host = host + n[0:index[len(index)-1]]
			n = n[index[len(index)-1]:]
		}
	}
	return host, n
}

func normalizeGitHostSpec(host string) string {
	s := strings.ToLower(host)
	if strings.Contains(s, "github.com") {
		if strings.Contains(s, "git@") || strings.Contains(s, "ssh:") {
			host = "git@github.com:"
		} else {
			host = "https://github.com/"
		}
	}
	if strings.HasPrefix(s, "git::") {
		host = strings.TrimLeft(s, "git::")
	}
	return host
}

// The format of Azure repo URL is documented
// https://docs.microsoft.com/en-us/azure/devops/repos/git/clone?view=vsts&tabs=visual-studio#clone_url
func isAzureHost(host string) bool {
	return strings.Contains(host, "dev.azure.com") ||
		strings.Contains(host, "visualstudio.com")
}

// The format of AWS repo URL is documented
// https://docs.aws.amazon.com/codecommit/latest/userguide/regions.html
func isAWSHost(host string) bool {
	return strings.Contains(host, "amazonaws.com")
}
