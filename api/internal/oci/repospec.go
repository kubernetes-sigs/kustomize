// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Used as a temporary non-empty occupant of the pullDir
// field, as something distinguishable from the empty string
// in various outputs (especially tests). Not using an
// actual directory name here, as that's a temporary directory
// with a unique name that isn't created until pull time.
const notPulled = filesys.ConfirmedDir("/notPulled")

// RepoSpec specifies an OCI repository and a tag
// TODO: and path therein?
type RepoSpec struct {
	// Raw, original spec, used to look for cycles.
	// TODO(monopole): Drop raw, use processed fields instead.
	raw string

	// Host, e.g. ghcr.io/
	Host string

	// RepoPath name (Remote repository),
	// e.g. kubernetes-sigs/kustomize
	RepoPath string

	// Dir is where the manifest is pulled to.
	Dir filesys.ConfirmedDir

	// Relative path in the repository, and in the pullDir,
	// to a Kustomization.
	KustRootPath string

	// Tag reference.
	Tag string

	// Digest
	Digest string

	// Timeout is the maximum duration allowed for execing git commands.
	Timeout time.Duration
}

// RepoSpec returns a string suitable for pulling with tools like oras.land, eg "oras pull {spec}".
// Note that this doesn't work with oci-layout hosts, as it requires a separate cli flag.
func (x *RepoSpec) PullSpec() string {
	pullSpec := x.Host + x.RepoPath

	if x.Tag != "" {
		pullSpec += tagSeparator + x.Tag
	}

	if x.Digest != "" {
		pullSpec += digestSeparator + x.Digest
	}

	return pullSpec
}

func (x *RepoSpec) PullDir() filesys.ConfirmedDir {
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

const (
	tagSeparator    = ":"
	digestSeparator = "@"
	pathSeparator   = "/" // do not use filepath.Separator, as this is a URL
	rootSeparator   = "//"
)

// NewRepoSpecFromURL parses OCI reference paths.
// From strings like oci://ghcr.io/someOrg/someRepository:someTag or
// oci://ghcr.io/someOrg/someRepository:sha256:<digest>, extract
// the different parts of URL, set into a RepoSpec object and return RepoSpec object.
// It MUST return an error if the input is not an oci-like URL, as this is used by some code paths
// to distinguish between local and remote paths.
//
// In particular, NewManifestSpecFromURL separates the URL used to pull the manifest from the
// elements Kustomize uses for other purposes (e.g. query params that turn into args, and
// the path to the kustomization root within the repo).
func NewRepoSpecFromURL(n string) (*RepoSpec, error) {
	repoSpec := &RepoSpec{raw: n, Dir: notPulled, Timeout: defaultTimeout}
	if filepath.IsAbs(n) {
		return nil, fmt.Errorf("uri looks like abs path: %s", n)
	}

	var err error

	// Parse the host (e.g. scheme, domain) segment.
	repoSpec.Host, n, err = extractHost(n)
	if err != nil {
		return nil, err
	}

	repoSpec.KustRootPath, n, err = extractRoot(n)
	if err != nil {
		return nil, err
	}

	repoSpec.Digest, n, err = extractDigest(n)
	if err != nil {
		return nil, err
	}

	repoSpec.Tag, repoSpec.RepoPath, err = extractTag(n)
	if err != nil {
		return nil, err
	}

	if len(repoSpec.RepoPath) == 0 {
		return nil, errors.Errorf("failed to parse repo path segment")
	}

	return repoSpec, nil
}

func extractRoot(n string) (string, string, error) {
	if rootIndex := strings.LastIndex(n, rootSeparator); rootIndex >= 0 {
		root := n[rootIndex+len(rootSeparator):]

		if len(root) == 0 {
			return "", "", errors.Errorf("failed to parse root path segment")
		}
		if kustRootPathExitsRepo(root) {
			return "", "", errors.Errorf("root path exits repo")
		}

		return root, n[:rootIndex], nil
	}

	return "", n, nil
}

func kustRootPathExitsRepo(kustRootPath string) bool {
	cleanedPath := filepath.Clean(strings.TrimPrefix(kustRootPath, string(filepath.Separator)))
	pathElements := strings.Split(cleanedPath, string(filepath.Separator))
	return len(pathElements) > 0 &&
		pathElements[0] == filesys.ParentDir
}

// Arbitrary, but non-infinite, timeout for running commands.
const defaultTimeout = 27 * time.Second

func extractDigest(n string) (string, string, error) {
	if digestIndex := strings.LastIndex(n, digestSeparator); digestIndex >= 0 {
		digest := n[digestIndex+len(digestSeparator):]
		// Digest is at least 8 characters, but we don't validate the entire schema here
		if len(digest) < 8 {
			return "", "", errors.Errorf("failed to parse digest")
		}

		return digest, n[:digestIndex], nil
	}

	// No digest
	return "", n, nil
}

func extractTag(n string) (string, string, error) {
	if tagIndex := strings.LastIndex(n, tagSeparator); tagIndex >= 0 {
		tag := n[tagIndex+len(tagSeparator):]

		if len(tag) == 0 {
			return "", "", errors.Errorf("failed to parse tag segment")
		}

		return tag, n[:tagIndex], nil
	}

	return "", n, nil
}

func extractHost(n string) (string, string, error) {
	scheme, n, err := extractScheme(n)

	if err != nil {
		return "", "", err
	}

	// Now that we have extracted a valid scheme, we can parse host itself.

	// The oci layout specifies a path to a local oci layout directory or archive.
	// Everything after the scheme is actually part of that path.
	if scheme == ociLayoutScheme {
		return ociLayoutScheme, n, nil
	}

	var host, rest = n, ""
	if sepIndex := findPathSeparator(n); sepIndex >= 0 {
		host, rest = n[:sepIndex+1], n[sepIndex+1:]
	}

	// Host is required, so do not concat the scheme and username if we didn't find one.
	if host == "" {
		return "", "", errors.Errorf("failed to parse host segment")
	}
	return host, rest, nil
}

const ociScheme = "oci://"
const ociLayoutScheme = "oci-layout://"

func extractScheme(s string) (string, string, error) {
	for _, prefix := range []string{ociScheme, ociLayoutScheme} {
		if rest, found := trimPrefixIgnoreCase(s, prefix); found {
			return prefix, rest, nil
		}
	}
	return "", "", fmt.Errorf("unsupported scheme")
}

// trimPrefixIgnoreCase returns the rest of s and true if prefix, ignoring case, prefixes s.
// Otherwise, trimPrefixIgnoreCase returns s and false.
func trimPrefixIgnoreCase(s, prefix string) (string, bool) {
	if len(prefix) <= len(s) && strings.ToLower(s[:len(prefix)]) == prefix {
		return s[len(prefix):], true
	}
	return s, false
}

func findPathSeparator(hostPath string) int {
	sepIndex := strings.Index(hostPath, pathSeparator)
	return sepIndex
}
