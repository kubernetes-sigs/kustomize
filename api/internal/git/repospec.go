// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Used as a temporary non-empty occupant of the cloneDir
// field, as something distinguishable from the empty string
// in various outputs (especially tests). Not using an
// actual directory name here, as that's a temporary directory
// with a unique name that isn't created until clone time.
const notCloned = filesys.ConfirmedDir("/notCloned")

// Schemes returns the schemes that repo urls may have
func Schemes() []string {
	return []string{"https://", "http://", "ssh://"}
}

// RepoSpec specifies a git repository and a branch and path therein.
type RepoSpec struct {
	// Raw, original spec, used to look for cycles.
	// TODO(monopole): Drop raw, use processed fields instead.
	raw string

	// Host, e.g. https://github.com/
	Host string

	// orgRepo name (organization/repoName),
	// e.g. kubernetes-sigs/kustomize
	OrgRepo string

	// Dir where the orgRepo is cloned to.
	Dir filesys.ConfirmedDir

	// Relative path in the repository, and in the cloneDir,
	// to a Kustomization.
	Path string

	// Branch or tag reference.
	Ref string

	// e.g. .git or empty in case of _git is present
	GitSuffix string

	// Submodules indicates whether or not to clone git submodules.
	Submodules bool

	// Timeout is the maximum duration allowed for execing git commands.
	Timeout time.Duration
}

// CloneSpec returns a string suitable for "git clone {spec}".
func (x *RepoSpec) CloneSpec() string {
	if isAzureHost(x.Host) || isAWSHost(x.Host) {
		return x.Host + x.OrgRepo
	}
	return x.Host + x.OrgRepo + x.GitSuffix
}

func (x *RepoSpec) CloneDir() filesys.ConfirmedDir {
	return x.Dir
}

func (x *RepoSpec) Raw() string {
	return x.raw
}

func (x *RepoSpec) AbsPath() string {
	return x.Dir.Join(x.Path)
}

func (x *RepoSpec) Cleaner(fSys filesys.FileSystem) func() error {
	return func() error { return fSys.RemoveAll(x.Dir.String()) }
}

// NewRepoSpecFromURL parses git-like urls.
// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the parts.
func NewRepoSpecFromURL(n string) (*RepoSpec, error) {
	if filepath.IsAbs(n) {
		return nil, fmt.Errorf("uri looks like abs path: %s", n)
	}
	repoSpec := parseGitURL(n)
	if repoSpec.Host == "" {
		return nil, errors.Errorf("url lacks host: %s", n)
	}
	if repoSpec.OrgRepo == "" {
		return nil, errors.Errorf("url lacks orgRepo: %s", n)
	}
	return repoSpec, nil
}

const (
	refQuery     = "?ref="
	gitSuffix    = ".git"
	gitDelimiter = "_git/"
)

// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the parts.
func parseGitURL(n string) *RepoSpec {
	repoSpec := &RepoSpec{raw: n, Dir: notCloned}

	if strings.Contains(n, gitDelimiter) {
		index := strings.Index(n, gitDelimiter)
		// Adding _git/ to host
		repoSpec.Host = normalizeGitHostSpec(n[:index+len(gitDelimiter)])
		// url before next /, though make sure before ? query string delimiter
		repoSpec.OrgRepo = strings.Split(strings.Split(n[index+len(gitDelimiter):], "/")[0], "?")[0]
		repoSpec.Path, repoSpec.Ref, repoSpec.Timeout, repoSpec.Submodules =
			peelQuery(n[index+len(gitDelimiter)+len(repoSpec.OrgRepo):])
		return repoSpec
	}

	repoSpec.Host, n = parseHostSpec(n)

	repoSpec.GitSuffix = gitSuffix
	if strings.Contains(n, gitSuffix) {
		index := strings.Index(n, gitSuffix)
		repoSpec.OrgRepo = n[0:index]
		n = n[index+len(gitSuffix):]
		// always try to include / in previous component
		if len(n) > 0 && n[0] == '/' {
			n = n[1:]
		}
	} else if strings.Contains(n, "/") {
		// there exist at least 2 components that we can interpret as org + repo
		i := strings.Index(n, "/")
		j := strings.Index(n[i+1:], "/")
		// only 2 components, so can use peelQuery to group everything before ? query string delimiter as OrgRepo
		if j == -1 {
			repoSpec.OrgRepo, repoSpec.Ref, repoSpec.Timeout, repoSpec.Submodules = peelQuery(n)
			return repoSpec
		}
		// extract first 2 components
		j += i + 1
		repoSpec.OrgRepo = n[:j]
		n = n[j+1:]
	}
	repoSpec.Path, repoSpec.Ref, repoSpec.Timeout, repoSpec.Submodules = peelQuery(n)
	return repoSpec
}

// Clone git submodules by default.
const defaultSubmodules = true

// Arbitrary, but non-infinite, timeout for running commands.
const defaultTimeout = 27 * time.Second

func peelQuery(arg string) (string, string, time.Duration, bool) {
	// Parse the given arg into a URL. In the event of a parse failure, return
	// our defaults.
	parsed, err := url.Parse(arg)
	if err != nil {
		return arg, "", defaultTimeout, defaultSubmodules
	}
	values := parsed.Query()

	// ref is the desired git ref to target. Can be specified by in a git URL
	// with ?ref=<string> or ?version=<string>, although ref takes precedence.
	ref := values.Get("version")
	if queryValue := values.Get("ref"); queryValue != "" {
		ref = queryValue
	}

	// depth is the desired git exec timeout. Can be specified by in a git URL
	// with ?timeout=<duration>.
	duration := defaultTimeout
	if queryValue := values.Get("timeout"); queryValue != "" {
		// Attempt to first parse as a number of integer seconds (like "61"),
		// and then attempt to parse as a suffixed duration (like "61s").
		if intValue, err := strconv.Atoi(queryValue); err == nil && intValue > 0 {
			duration = time.Duration(intValue) * time.Second
		} else if durationValue, err := time.ParseDuration(queryValue); err == nil && durationValue > 0 {
			duration = durationValue
		}
	}

	// submodules indicates if git submodule cloning is desired. Can be
	// specified by in a git URL with ?submodules=<bool>.
	submodules := defaultSubmodules
	if queryValue := values.Get("submodules"); queryValue != "" {
		if boolValue, err := strconv.ParseBool(queryValue); err == nil {
			submodules = boolValue
		}
	}

	return parsed.Path, ref, duration, submodules
}

func parseHostSpec(n string) (string, string) {
	var host string
	// TODO(annasong): handle ports, non-github ssh with userinfo, github.com sub-domains
	// Start accumulating the host part.
	for _, p := range []string{
		// Order matters here.
		"git::", "gh:", "ssh://", "https://", "http://",
		"git@", "github.com:", "github.com/"} {
		if len(p) < len(n) && strings.ToLower(n[:len(p)]) == p {
			n = n[len(p):]
			host += p
		}
	}
	// ssh relative path
	if host == "git@" {
		var i int
		// git orgRepo delimiter for ssh relative path
		j := strings.Index(n, ":")
		// orgRepo delimiter only valid if not preceded by /
		k := strings.Index(n, "/")
		// orgRepo delimiter exists, so extend host
		if j > -1 && (k == -1 || k > j) {
			i = j
			// should try to group / with previous component
			// this is possible if git url username expansion follows
			if k == i+1 {
				i = k
			}
			host += n[:i+1]
			n = n[i+1:]
		}
		return host, n
	}

	// If host is a http(s) or ssh URL, grab the domain part.
	for _, p := range Schemes() {
		if strings.HasSuffix(host, p) {
			i := strings.Index(n, "/")
			if i > -1 {
				host += n[0 : i+1]
				n = n[i+1:]
			}
			break
		}
	}

	return normalizeGitHostSpec(host), n
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
		host = strings.TrimPrefix(s, "git::")
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
