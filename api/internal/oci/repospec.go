// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package oci supports loading kustomizations from OCI artifacts.
package oci

import (
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Used as a temporary non-empty occupant of the Dir
// field, as something distinguishable from the empty string
// in various outputs (especially tests). Not using an
// actual directory name here, as that's a temporary directory
// with a unique name that isn't created until pull time.
const notPulled = filesys.ConfirmedDir("/notPulled")

// RepoSpec specifies an OCI artifact and a path therein.
type RepoSpec struct {
	// Raw, original spec, used to look for cycles.
	raw string

	// Reference is the parsed registry/repository[:tag|@digest].
	Reference name.Reference

	// Dir is where the artifact is extracted to.
	Dir filesys.ConfirmedDir

	// Relative path in the artifact, and in Dir,
	// to a Kustomization.
	KustRootPath string
}

// PullSpec returns the fully qualified reference of the artifact,
// suitable for pulling with tools like crane or oras.
func (x *RepoSpec) PullSpec() string {
	return x.Reference.Name()
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
	return func() error { return errors.Wrap(fSys.RemoveAll(x.Dir.String())) }
}

const (
	ociScheme     = "oci://"
	defaultTag    = "latest"
	pathSeparator = "/" // do not use filepath.Separator, as this is a URL
	rootDelimiter = "//"
)

// IsOciURL returns whether the path has the oci:// scheme,
// ignoring case, i.e. whether it refers to an OCI artifact.
func IsOciURL(path string) bool {
	_, found := trimPrefixIgnoreCase(path, ociScheme)
	return found
}

// NewRepoSpecFromURL parses OCI artifact URLs.
// From strings like oci://ghcr.io/someOrg/someRepo:someTag//path/to/root or
// oci://ghcr.io/someOrg/someRepo@sha256:<digest>, extract the artifact
// reference and the optional kustomization root path within the artifact,
// set them into a RepoSpec object and return the RepoSpec object.
// The tag defaults to "latest" when neither a tag nor a digest is given.
// It MUST return an error if the input is not an oci:// URL, as this is used
// by some code paths to distinguish between local and remote paths.
func NewRepoSpecFromURL(n string) (*RepoSpec, error) {
	rest, found := trimPrefixIgnoreCase(n, ociScheme)
	if !found {
		return nil, errors.Errorf("uri does not have the %s scheme: %s", ociScheme, n)
	}

	ref, kustRootPath, err := splitRootPath(rest)
	if err != nil {
		return nil, err
	}

	reference, err := name.ParseReference(ref,
		name.WithDefaultRegistry(""),
		name.WithDefaultTag(defaultTag))
	if err != nil {
		return nil, errors.WrapPrefixf(err,
			"invalid oci artifact reference %q (the kustomization root path must be separated by %s)",
			ref, rootDelimiter)
	}
	if reference.Context().RegistryStr() == "" {
		return nil, errors.Errorf("oci artifact reference %q must include the registry host", ref)
	}

	return &RepoSpec{
		raw:          n,
		Reference:    reference,
		Dir:          notPulled,
		KustRootPath: kustRootPath,
	}, nil
}

// splitRootPath splits the artifact reference from the path to the
// kustomization root within the artifact. By convention a double-slash
// separates the two, e.g. ghcr.io/org/repo:tag//path/to/root. The
// delimiter is not a real path element, so it is not preserved.
func splitRootPath(n string) (string, string, error) {
	rootIdx := strings.Index(n, rootDelimiter)
	if rootIdx < 0 {
		return n, "", nil
	}
	ref, kustRootPath := n[:rootIdx], n[rootIdx+len(rootDelimiter):]
	if kustRootPathExitsRepo(kustRootPath) {
		return "", "", errors.Errorf("url path exits artifact: %s", n)
	}
	return ref, strings.TrimPrefix(kustRootPath, pathSeparator), nil
}

func kustRootPathExitsRepo(kustRootPath string) bool {
	cleanedPath := filepath.Clean(strings.TrimPrefix(kustRootPath, string(filepath.Separator)))
	pathElements := strings.Split(cleanedPath, string(filepath.Separator))
	return len(pathElements) > 0 &&
		pathElements[0] == filesys.ParentDir
}

// trimPrefixIgnoreCase returns the rest of s and true if prefix, ignoring case, prefixes s.
// Otherwise, trimPrefixIgnoreCase returns s and false.
func trimPrefixIgnoreCase(s, prefix string) (string, bool) {
	if len(prefix) <= len(s) && strings.ToLower(s[:len(prefix)]) == prefix {
		return s[len(prefix):], true
	}
	return s, false
}
