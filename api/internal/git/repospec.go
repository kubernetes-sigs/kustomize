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

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Used as a temporary non-empty occupant of the cloneDir
// field, as something distinguishable from the empty string
// in various outputs (especially tests). Not using an
// actual directory name here, as that's a temporary directory
// with a unique name that isn't created until clone time.
const notCloned = filesys.ConfirmedDir("/notCloned")

// RepoSpec specifies a git repository and a branch and path therein.
type RepoSpec struct {
	// Raw, original spec, used to look for cycles.
	// TODO(monopole): Drop raw, use processed fields instead.
	raw string

	// Host, e.g. https://github.com/
	Host string

	// RepoPath name (Path to repository),
	// e.g. kubernetes-sigs/kustomize
	RepoPath string

	// Dir is where the repository is cloned to.
	Dir filesys.ConfirmedDir

	// Relative path in the repository, and in the cloneDir,
	// to a Kustomization.
	KustRootPath string

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
		return x.Host + x.RepoPath
	}
	return x.Host + x.RepoPath + x.GitSuffix
}

func (x *RepoSpec) CloneDir() filesys.ConfirmedDir {
	return x.Dir
}

func (x *RepoSpec) Raw() string {
	return x.raw
}

func (x *RepoSpec) AbsPath() string {
	return x.Dir.Join(x.KustRootPath)
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
	repoSpecVal := parseGitURL(n)
	if repoSpecVal.RepoPath == "" {
		return nil, fmt.Errorf("url lacks repoPath: %s", n)
	}
	if repoSpecVal.Host == "" {
		return nil, fmt.Errorf("url lacks host: %s", n)
	}
	cleanedPath := filepath.Clean(strings.TrimPrefix(repoSpecVal.KustRootPath, string(filepath.Separator)))
	if pathElements := strings.Split(cleanedPath, string(filepath.Separator)); len(pathElements) > 0 &&
		pathElements[0] == filesys.ParentDir {
		return nil, fmt.Errorf("url path exits repo: %s", n)
	}
	return repoSpecVal, nil
}

const (
	refQuery     = "?ref="
	gitSuffix    = ".git"
	gitDelimiter = "_git/"
)

// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the different parts of URL , set into a RepoSpec object and return RepoSpec object.
func parseGitURL(n string) *RepoSpec {
	repoSpec := &RepoSpec{raw: n, Dir: notCloned, Timeout: defaultTimeout, Submodules: defaultSubmodules}
	// parse query first
	// safe because according to rfc3986: ? only allowed in query
	// and not recognized %-encoded
	beforeQuery, query, _ := strings.Cut(n, "?")
	n = beforeQuery
	// if no query, defaults returned
	repoSpec.Ref, repoSpec.Timeout, repoSpec.Submodules = parseQuery(query)

	if strings.Contains(n, gitDelimiter) {
		index := strings.Index(n, gitDelimiter)
		// Adding _git/ to host
		repoSpec.Host = normalizeGitHostSpec(n[:index+len(gitDelimiter)])
		repoSpec.RepoPath = strings.Split(n[index+len(gitDelimiter):], "/")[0]
		repoSpec.KustRootPath = parsePath(n[index+len(gitDelimiter)+len(repoSpec.RepoPath):])
		return repoSpec
	}
	repoSpec.Host, n = parseHostSpec(n)
	isLocal := strings.HasPrefix(repoSpec.Host, "file://")
	if !isLocal {
		repoSpec.GitSuffix = gitSuffix
	}
	if strings.Contains(n, gitSuffix) {
		repoSpec.GitSuffix = gitSuffix
		index := strings.Index(n, gitSuffix)
		repoSpec.RepoPath = n[0:index]
		n = n[index+len(gitSuffix):]
		if len(n) > 0 && n[0] == '/' {
			n = n[1:]
		}
		repoSpec.KustRootPath = parsePath(n)
		return repoSpec
	}

	if isLocal {
		if idx := strings.Index(n, "//"); idx > 0 {
			repoSpec.RepoPath = n[:idx]
			n = n[idx+2:]
			repoSpec.KustRootPath = parsePath(n)
			return repoSpec
		}
		repoSpec.RepoPath = parsePath(n)
		return repoSpec
	}

	i := strings.Index(n, "/")
	if i < 1 {
		repoSpec.KustRootPath = parsePath(n)
		return repoSpec
	}
	j := strings.Index(n[i+1:], "/")
	if j >= 0 {
		j += i + 1
		repoSpec.RepoPath = n[:j]
		repoSpec.KustRootPath = parsePath(n[j+1:])
		return repoSpec
	}
	repoSpec.KustRootPath = ""
	repoSpec.RepoPath = parsePath(n)
	return repoSpec
}

// Clone git submodules by default.
const defaultSubmodules = true

// Arbitrary, but non-infinite, timeout for running commands.
const defaultTimeout = 27 * time.Second

func parseQuery(query string) (string, time.Duration, bool) {
	values, err := url.ParseQuery(query)
	// in event of parse failure, return defaults
	if err != nil {
		return "", defaultTimeout, defaultSubmodules
	}

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

	return ref, duration, submodules
}

func parsePath(n string) string {
	parsed, err := url.Parse(n)
	// TODO(annasong): decide how to handle error, i.e. return error, empty string, etc.
	if err != nil {
		return n
	}
	return parsed.Path
}

func parseHostSpec(n string) (string, string) {
	// We used to use go-getter to handle our urls: https://github.com/hashicorp/go-getter.
	// This prefix signaled go-getter to use the git protocol to fetch the url's contents.
	// We still accept this prefix.
	n, _ = trimPrefixIgnoreCase(n, "git::")
	// We support gh: assuming authors use it as a github shorthand, specified in .gitconfig.
	if rest, found := trimPrefixIgnoreCase(n, "gh:"); found {
		return "gh:", rest
	}
	scheme, rest, _ := strings.Cut(n, "://")
	switch scheme = strings.ToLower(scheme); scheme {
	// no scheme
	case strings.ToLower(n):
		scheme = ""
		rest = n
	// The file protocol specifies an absolute path to a local git repo. There is no host.
	case "file":
		return "file://", rest
	case "https", "http", "ssh":
		scheme += "://"
	// We either
	//   1. do not support said scheme or
	//   2. found a part of the path because there is no scheme.
	// Instead of determining the case, we try to match the host as if under the 2nd case.
	// If we are actually under the first, host matching will fail.
	default:
		scheme = ""
		rest = n
	}
	return matchHost(scheme, rest)
}

// trimPrefixIgnoreCase returns the rest of s and true if prefix, ignoring case, prefixes s.
// Otherwise, trimPrefixIgnoreCase returns s and false.
func trimPrefixIgnoreCase(s, prefix string) (string, bool) {
	if len(prefix) <= len(s) && strings.ToLower(s[:len(prefix)]) == prefix {
		return s[len(prefix):], true
	}
	return s, false
}

// matchHost returns a host that kustomize recognizes and the rest of s given scheme and url s.
// scheme can be any of http;//, https://, ssh://, or empty.
func matchHost(scheme, s string) (host, rest string) {
	var isSCP bool
	if s, isSCP = trimPrefixIgnoreCase(s, "git@"); isSCP {
		host = "git@"
	}
	const httpGithub = "https://github.com/"
	const scpGithub = "git@github.com:"
	var normalized, separator string
	switch scheme {
	case "":
		if isSCP {
			normalized = scpGithub
			separator = ":"
		} else {
			normalized = httpGithub
		}
	case "https://", "http://":
		normalized = httpGithub
		separator = "/"
	case "ssh://":
		normalized = scpGithub
		separator = "/"
	}
	for _, builtin := range []string{"github.com:", "github.com/"} {
		rest, found := trimPrefixIgnoreCase(s, builtin)
		if found {
			return normalized, rest
		}
	}
	i := strings.Index(s, separator)
	// There is no host if the separator was not found or the separator delimits an empty
	// host. Note that this will happen if the separator is empty.
	if i <= 0 {
		return "", s
	}
	if separator == ":" {
		// The colon acts as a delimiter for scp protocol only if not prefixed by '/'.
		if slashIndex := strings.Index(s, "/"); slashIndex != -1 && slashIndex < i {
			return "", s
		}
	}
	return scheme + host + s[:i+1], s[i+1:]
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
