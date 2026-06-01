// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
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

	tag string

	digest string

	// Dir is where the manifest is pulled to.
	Dir filesys.ConfirmedDir

	// If a remote artifact, the reference to the remote artifact
	Reference name.Reference

	// Relative path in the repository, and in the pullDir,
	// to a Kustomization.
	KustRootPath string

	// Timeout is the maximum duration allowed for execing git commands.
	Timeout time.Duration
}

// RepoSpec returns a string suitable for pulling with tools like oras.land, eg "oras pull {spec}".
// Note that this doesn't work with oci-layout hosts, as it requires a separate cli flag.
func (x *RepoSpec) PullSpec() string {
	return x.Reference.String()
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

	n, err := trimScheme(n)
	if err != nil {
		return nil, err
	}

	repoSpec.KustRootPath, n, err = extractRoot(n)
	if err != nil {
		return nil, err
	}

	repoSpec.Reference, err = name.ParseReference(n,
		name.WithDefaultRegistry(""),
		name.WithDefaultTag("latest"),
	)
	if err != nil {
		return nil, err
	}
	if repoSpec.Reference.Context().Registry.Name() == "" || repoSpec.Reference.Context().RepositoryStr() == "" {
		return nil, fmt.Errorf("invalid reference: missing registry or repository")
	}

	return repoSpec, nil
}

func extractRoot(n string) (string, string, error) {
	if rootIndex := strings.LastIndex(n, rootSeparator); rootIndex >= 0 {
		root := n[rootIndex+len(rootSeparator):]

		if root == "" {
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

const ociScheme = "oci://"

func trimScheme(s string) (string, error) {
	if len(ociScheme) <= len(s) && strings.ToLower(s[:len(ociScheme)]) == ociScheme {
		return s[len(ociScheme):], nil
	}
	return "", fmt.Errorf("unsupported scheme")
}
