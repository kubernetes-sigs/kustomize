// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
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
	rs, err := parseGitURL(n)
	if err != nil {
		return nil, err
	}
	if rs.OrgRepo == "" {
		return nil, fmt.Errorf("url lacks orgRepo: %s", n)
	}
	if rs.Host == "" {
		return nil, fmt.Errorf("url lacks host: %s", n)
	}
	return rs, nil
}

const (
	refQuery     = "?ref="
	gitSuffix    = ".git"
	gitDelimiter = "_git/"
)

// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the parts.
func parseGitURL(n string) (*RepoSpec, error) {
	var err error
	rs := &RepoSpec{raw: n, Dir: notCloned}
	if strings.Contains(n, gitDelimiter) {
		index := strings.Index(n, gitDelimiter)
		// Adding _git/ to host
		rs.Host, err = normalizeGitHostSpec(n[:index+len(gitDelimiter)])
		if err != nil {
			return nil, err
		}
		rs.OrgRepo = strings.Split(strings.Split(n[index+len(gitDelimiter):], "/")[0], "?")[0]
		rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n[index+len(gitDelimiter)+len(rs.OrgRepo):])
		return rs, nil
	}
	rs.Host, n, err = parseHostSpec(n)
	if err != nil {
		return nil, err
	}
	isLocal := strings.HasPrefix(rs.Host, "file://")
	if !isLocal {
		rs.GitSuffix = gitSuffix
	}
	if strings.Contains(n, gitSuffix) {
		rs.GitSuffix = gitSuffix
		index := strings.Index(n, gitSuffix)
		rs.OrgRepo = n[0:index]
		n = n[index+len(gitSuffix):]
		if len(n) > 0 && n[0] == '/' {
			n = n[1:]
		}
		rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n)
		return rs, nil
	}

	if isLocal {
		if idx := strings.Index(n, "//"); idx > 0 {
			rs.OrgRepo = n[:idx]
			n = n[idx+2:]
			rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n)
			return rs, nil
		}
		rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n)
		rs.OrgRepo = rs.Path
		rs.Path = ""
		return rs, nil
	}

	i := strings.Index(n, "/")
	if i < 1 {
		rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n)
		return rs, nil
	}
	j := strings.Index(n[i+1:], "/")
	if j >= 0 {
		j += i + 1
		rs.OrgRepo = n[:j]
		rs.Path, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n[j+1:])
		return rs, nil
	}
	rs.Path = ""
	rs.OrgRepo, rs.Ref, rs.Timeout, rs.Submodules = peelQuery(n)
	return rs, nil
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

var userRegexp = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9-]*)@`)

func parseHostSpec(n string) (string, string, error) {
	var host string
	consumeHostStrings := func(parts []string) {
		for _, p := range parts {
			if len(p) < len(n) && strings.ToLower(n[:len(p)]) == p {
				n = n[len(p):]
				host += p
			}
		}
	}
	// Start accumulating the host part.
	// Order matters here.
	consumeHostStrings([]string{"git::", "gh:", "ssh://", "https://", "http://", "file://"})
	if p := userRegexp.FindString(n); p != "" {
		n = n[len(p):]
		host += p
	}
	consumeHostStrings([]string{"github.com:", "github.com/"})
	if host == "git@" {
		i := strings.Index(n, "/")
		if i > -1 {
			host += n[:i+1]
			n = n[i+1:]
		} else {
			i = strings.Index(n, ":")
			if i > -1 {
				host += n[:i+1]
				n = n[i+1:]
			}
		}
		return host, n, nil
	}

	// If host is a http(s) or ssh URL, grab the domain part.
	for _, p := range []string{
		"ssh://", "https://", "http://"} {
		if strings.HasSuffix(host, p) {
			i := strings.Index(n, "/")
			if i > -1 {
				host += n[0 : i+1]
				n = n[i+1:]
			}
			break
		}
	}

	host, err := normalizeGitHostSpec(host)
	return host, n, err
}

var githubRegexp = regexp.MustCompile(`^(?:ssh://)?([a-zA-Z][a-zA-Z0-9-]*)@(github.com[:/]?)`)

func normalizeGitHostSpec(host string) (string, error) {
	s := strings.ToLower(host)

	// The git:: syntax is meant to force the Git protocol (separate from SSH
	// and HTTPS), but we drop it here, to preserve past behavior.
	isGitProtocol := strings.HasPrefix(s, "git::")
	if isGitProtocol {
		host = strings.TrimPrefix(s, "git::")
	}

	// Special treatment for github.com
	if strings.Contains(host, "github.com") {
		m := githubRegexp.FindStringSubmatch(host)
		if m == nil {
			return "https://github.com/", nil
		}
		userName, realHost := m[1], m[2]

		if realHost == "github.com/" {
			realHost = "github.com:"
		}

		const gitUser = "git"
		isGitUser := userName == gitUser || userName == ""
		if userName == "" {
			userName = gitUser
		}

		switch {
		case isGitProtocol && !isGitUser:
			return "", fmt.Errorf("git protocol on github.com only allows git@ user")
		case isGitProtocol:
			return "git@github.com:", nil
		default:
			return fmt.Sprintf("%s@%s", userName, realHost), nil
		}
	}
	return host, nil
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
