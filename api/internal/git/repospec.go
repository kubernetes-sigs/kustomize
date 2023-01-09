// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"log"
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
	repoSpec.Host, n = extractHost(n)
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

func extractHost(n string) (string, string) {
	n = ignoreForcedGitProtocol(n)
	scheme, n := extractScheme(n)
	username, n := extractUsername(n)
	stdGithub := isStandardGithubHost(n)
	acceptSCP := acceptSCPStyle(scheme, username, stdGithub)

	// Validate the username and scheme before attempting host/path parsing, because if the parsing
	// so far has not succeeded, we will not be able to extract the host and path correctly.
	if err := validateUsernameAndScheme(username, scheme, acceptSCP); err != nil {
		// TODO: return this error instead.
		return "", n
	}

	// Now that we have extracted a valid scheme+username, we can parse host itself.

	// The file protocol specifies an absolute path to a local git repo.
	// Everything after the scheme (including any 'username' we found) is actually part of that path.
	if scheme == "file://" {
		return scheme, username + n
	}
	sepIndex := findPathSeparator(n, acceptSCP)
	host, rest := n[:sepIndex+1], n[sepIndex+1:]

	// Github URLs are strictly normalized in a way that may discard scheme and username components.
	if stdGithub {
		scheme, username, host = normalizeGithubHostParts(scheme, username)
	}

	// Host is required, so do not concat the scheme and username if we didn't find one.
	if host == "" {
		// TODO: This should return an error.
		return "", n
	}
	return scheme + username + host, rest
}

// ignoreForcedGitProtocol strips the "git::" prefix from URLs.
// We used to use go-getter to handle our urls: https://github.com/hashicorp/go-getter.
// The git:: prefix signaled go-getter to use the git protocol to fetch the url's contents.
// We silently strip this prefix to allow these go-getter-style urls to continue to work,
// although the git protocol (which is insecure and unsupported on many platforms, including Github)
// will not actually be used as intended.
func ignoreForcedGitProtocol(n string) string {
	n, found := trimPrefixIgnoreCase(n, "git::")
	if found {
		log.Println("Warning: Forcing the git protocol using the 'git::' URL prefix is not supported. " +
			"Kustomize currently strips this invalid prefix, but will stop doing so in a future release. " +
			"Please remove the 'git::' prefix from your configuration.")
	}
	return n
}

// acceptSCPStyle returns true if the scheme and username indicate potential use of an SCP-style URL.
// With this style, the scheme is not explicit and the path is delimited by a colon.
// Strictly speaking the username is optional in SCP-like syntax, but Kustomize has always
// required it for non-Github URLs.
// Example: user@host.xz:path/to/repo.git/
func acceptSCPStyle(scheme, username string, isGithubURL bool) bool {
	return scheme == "" && (username != "" || isGithubURL)
}

func validateUsernameAndScheme(username, scheme string, acceptSCPStyle bool) error {
	// see https://git-scm.com/docs/git-fetch#_git_urls for info relevant to these validations
	switch scheme {
	case "":
		// Empty scheme is only ok if it's a Github URL or if it looks like SCP-style syntax
		if !acceptSCPStyle {
			return fmt.Errorf("no username or scheme found")
		}
	case "ssh://":
		// usernames are optional for ssh protocol
		return nil
	case "file://":
		// everything following the scheme in the file protocol is a path on the local filesystem,
		// which may contain arbitrary characters (theoretically including `@`, which we'd mistake for a username)
		return nil
	case "https://", "http://":
		// usernames are not supported by the http protocol
		if username != "" {
			return fmt.Errorf("username %q specified, but %s does not support usernames", username, scheme)
		}
	default:
		// At time of writing, we should never end up here because we do not parse out
		// unsupported schemes to begin with.
		return fmt.Errorf("unsupported scheme %q", scheme)
	}
	return nil
}

func extractScheme(s string) (string, string) {
	for _, prefix := range []string{"ssh://", "https://", "http://", "file://"} {
		if rest, found := trimPrefixIgnoreCase(s, prefix); found {
			return prefix, rest
		}
	}
	return "", s
}

func extractUsername(s string) (string, string) {
	if trimmed, found := trimPrefixIgnoreCase(s, gitUsername); found {
		return gitUsername, trimmed
	}
	return "", s
}

func isStandardGithubHost(s string) bool {
	lowerCased := strings.ToLower(s)
	return strings.HasPrefix(lowerCased, "github.com/") ||
		strings.HasPrefix(lowerCased, "github.com:")
}

// trimPrefixIgnoreCase returns the rest of s and true if prefix, ignoring case, prefixes s.
// Otherwise, trimPrefixIgnoreCase returns s and false.
func trimPrefixIgnoreCase(s, prefix string) (string, bool) {
	if len(prefix) <= len(s) && strings.ToLower(s[:len(prefix)]) == prefix {
		return s[len(prefix):], true
	}
	return s, false
}

func findPathSeparator(hostPath string, acceptSCP bool) int {
	sepIndex := strings.Index(hostPath, "/")
	if acceptSCP {
		// The colon acts as a delimiter in scp-style ssh URLs only if not prefixed by '/'.
		if colonIndex := strings.Index(hostPath, ":"); colonIndex > 0 && colonIndex < sepIndex {
			sepIndex = colonIndex
		}
	}
	return sepIndex
}

const normalizedHTTPGithub = "https://github.com/"
const gitUsername = "git@"
const normalizedSCPGithub = gitUsername + "github.com:"

func normalizeGithubHostParts(scheme, username string) (string, string, string) {
	if strings.HasPrefix(scheme, "ssh://") || username != "" {
		return "", username, "github.com:"
	}
	return "https://", "", "github.com/"
}

func normalizeGitHostSpec(host string) string {
	s := strings.ToLower(host)
	if strings.Contains(s, "github.com") {
		if strings.Contains(s, gitUsername) || strings.Contains(s, "ssh:") {
			host = normalizedSCPGithub
		} else {
			host = normalizedHTTPGithub
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
