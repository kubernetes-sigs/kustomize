// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// IsRemoteFile returns whether path has a url scheme that kustomize allows for
// remote files. See https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md
func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// FileLoader is a kustomization's interface to files.
//
// The directory in which a kustomization file sits
// is referred to below as the kustomization's _root_.
//
// An instance of fileLoader has an immutable root,
// and offers a `New` method returning a new loader
// with a new root.
//
// A kustomization file refers to two kinds of files:
//
// * supplemental data paths
//
//	`Load` is used to visit these paths.
//
//	These paths refer to resources, patches,
//	data for ConfigMaps and Secrets, etc.
//
//	The loadRestrictor may disallow certain paths
//	or classes of paths.
//
// * bases (other kustomizations)
//
//	`New` is used to load bases.
//
//	A base can be either a remote git repo or OCI
//	artifact URL, or a directory specified relative
//	to the current root. In the former case, the
//	repo is locally cloned, or the artifact pulled,
//	and the new loader is rooted on a path in that
//	copy.
//
//	As loaders create new loaders, a root history
//	is established, and used to disallow:
//
//	- A base that is a repository that, in turn,
//	  specifies a base repository seen previously
//	  in the loading stack (a cycle).
//
//	- An overlay depending on a base positioned at
//	  or above it.  I.e. '../foo' is OK, but '.',
//	  '..', '../..', etc. are disallowed.  Allowing
//	  such a base has no advantages and encourages
//	  cycles, particularly if some future change
//	  were to introduce globbing to file
//	  specifications in the kustomization file.
//
// These restrictions assure that kustomizations
// are self-contained and relocatable, and impose
// some safety when relying on remote kustomizations,
// e.g. a remotely loaded ConfigMap generator specified
// to read from /etc/passwd will fail.
type FileLoader struct {
	// Loader that spawned this loader.
	// Used to avoid cycles.
	referrer *FileLoader

	// An absolute, cleaned path to a directory.
	// The Load function will read non-absolute
	// paths relative to this directory.
	root filesys.ConfirmedDir

	// Restricts behavior of Load function.
	loadRestrictor LoadRestrictorFunc

	// If this is non-nil, the files were obtained
	// from the given git repo or OCI artifact.
	remote remoteSpec

	// File system utilities.
	fSys filesys.FileSystem

	// Used to load from HTTP
	http *http.Client

	// Used to clone repositories.
	cloner git.Cloner

	// Used to pull OCI artifacts.
	puller oci.Puller

	// If this is non-nil, the clones and pulls
	// shared by the loaders of a build.
	remotes *remoteCache

	// Used to clean up, as needed.
	cleaner func() error
}

// remoteSpec is a remote root, a git repo or an OCI artifact, as the
// loaders see it once fetched. *git.RepoSpec and *oci.RepoSpec
// implement it.
type remoteSpec interface {
	// Raw returns the reference as written; cycle checks compare it.
	Raw() string
	// FetchDir returns the directory holding the clone or pull.
	FetchDir() filesys.ConfirmedDir
	// AbsPath returns the root of the kustomization in that directory.
	AbsPath() string
}

// remoteKind returns how errors name the kind of remote: what holds
// such remotes, and what one is.
func remoteKind(remote remoteSpec) (fetched, unit string) {
	if _, isArtifact := remote.(*oci.RepoSpec); isArtifact {
		return "pulled oci artifacts", "artifact"
	}
	return "cloned git repos", "repo"
}

// Repo returns the absolute path to the repo that contains Root if this fileLoader was created from a url
// or the empty string otherwise.
func (fl *FileLoader) Repo() string {
	if fl.remote != nil {
		return fl.remote.FetchDir().String()
	}
	return ""
}

// Root returns the absolute path that is prepended to any
// relative paths used in Load.
func (fl *FileLoader) Root() string {
	return fl.root.String()
}

func NewLoaderOrDie(
	lr LoadRestrictorFunc,
	fSys filesys.FileSystem, path string) *FileLoader {
	root, err := filesys.ConfirmDir(fSys, path)
	if err != nil {
		log.Fatalf("unable to make loader at '%s'; %v", path, err)
	}
	return newLoaderAtConfirmedDir(
		lr, root, fSys, nil, git.ClonerUsingGitExec, oci.PullerUsingRegistry, nil)
}

// newLoaderAtConfirmedDir returns a new FileLoader with given root.
func newLoaderAtConfirmedDir(
	lr LoadRestrictorFunc,
	root filesys.ConfirmedDir, fSys filesys.FileSystem,
	referrer *FileLoader, cloner git.Cloner, puller oci.Puller,
	remotes *remoteCache) *FileLoader {
	return &FileLoader{
		loadRestrictor: lr,
		root:           root,
		referrer:       referrer,
		fSys:           fSys,
		cloner:         cloner,
		puller:         puller,
		remotes:        remotes,
		cleaner:        func() error { return nil },
	}
}

// hasAmbiguousPathSegment reports whether path contains a segment with ".."
// embedded in it (not ".." itself). This pattern results from YAML indentation
// errors where two list entries merge into one scalar, e.g.:
//
//	resources:
//	- ../../base
//	 - ../../shared/prod
//
// YAML parses this as "../../base - ../../shared/prod", producing segment
// "base - .." which filepath.Clean absorbs, silently resolving to an
// unintended directory.
func hasAmbiguousPathSegment(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part != "" && part != "." && part != ".." && strings.Contains(part, "..") {
			return true
		}
	}
	return false
}

// New returns a new Loader, rooted relative to current loader,
// or rooted in a temp directory holding a git repo clone
// or a pulled OCI artifact.
func (fl *FileLoader) New(path string) (ifc.Loader, error) {
	if path == "" {
		return nil, errors.Errorf("new root cannot be empty")
	}

	repoSpec, err := git.NewRepoSpecFromURL(path)
	if err == nil {
		// Treat this as git repo clone request.
		if err = fl.errIfRepoCycle(repoSpec); err != nil {
			return nil, err
		}
		return newLoaderAtGitClone(
			repoSpec, fl.fSys, fl, fl.cloner, fl.puller, fl.remotes)
	}

	ociSpec, err := oci.NewRepoSpecFromURL(path)
	if err == nil {
		// Treat this as OCI artifact pull request.
		if err = fl.errIfRepoCycle(ociSpec); err != nil {
			return nil, err
		}
		return newLoaderAtOciPull(
			ociSpec, fl.fSys, fl, fl.cloner, fl.puller, fl.remotes)
	}
	if oci.IsOciURL(path) {
		// An oci:// URL that does not parse is not a local path either.
		return nil, errors.Wrap(err)
	}

	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("new root '%s' cannot be absolute", path)
	}
	if hasAmbiguousPathSegment(path) {
		return nil, fmt.Errorf(
			"path %q is ambiguous: resolves to %q after normalization, "+
				"which is likely not the intended target; check for YAML "+
				"indentation errors in your kustomization file",
			path, filepath.Clean(path))
	}
	root, err := filesys.ConfirmDir(fl.fSys, fl.root.Join(path))
	if err != nil {
		return nil, errors.WrapPrefixf(err, "%s", ErrRtNotDir.Error())
	}
	if err = fl.errIfContainmentViolation(root); err != nil {
		return nil, err
	}
	if err = fl.errIfArgEqualOrHigher(root); err != nil {
		return nil, err
	}
	return newLoaderAtConfirmedDir(
		fl.loadRestrictor, root, fl.fSys, fl, fl.cloner, fl.puller, fl.remotes), nil
}

// newLoaderAtGitClone returns a new Loader pinned to a temporary
// directory holding a cloned git repo, from the remotes of the build
// if they include it.
func newLoaderAtGitClone(
	repoSpec *git.RepoSpec, fSys filesys.FileSystem,
	referrer *FileLoader, cloner git.Cloner, puller oci.Puller,
	remotes *remoteCache) (ifc.Loader, error) {
	cleaner, err := remotes.fetch(gitCacheKey(repoSpec), &repoSpec.Dir,
		func() error { return cloner(repoSpec) }, repoSpec.Cleaner(fSys))
	if err != nil {
		return nil, err
	}
	return newLoaderAtRemote(repoSpec, cleaner, fSys, referrer, cloner, puller, remotes)
}

// newLoaderAtOciPull returns a new Loader pinned to a temporary
// directory holding a pulled OCI artifact, from the remotes of the
// build if they include it.
func newLoaderAtOciPull(
	repoSpec *oci.RepoSpec, fSys filesys.FileSystem,
	referrer *FileLoader, cloner git.Cloner, puller oci.Puller,
	remotes *remoteCache) (ifc.Loader, error) {
	var client *http.Client
	if referrer != nil {
		client = referrer.http
	}
	cleaner, err := remotes.fetch(ociCacheKey(repoSpec), &repoSpec.Dir,
		func() error { return puller(repoSpec, fSys, client) }, repoSpec.Cleaner(fSys))
	if err != nil {
		return nil, err
	}
	return newLoaderAtRemote(repoSpec, cleaner, fSys, referrer, cloner, puller, remotes)
}

// newLoaderAtRemote returns a new Loader at the kustomization root of a
// fetched git repo or OCI artifact, which cleaner removes as needed.
func newLoaderAtRemote(
	remote remoteSpec, cleaner func() error, fSys filesys.FileSystem,
	referrer *FileLoader, cloner git.Cloner, puller oci.Puller,
	remotes *remoteCache) (ifc.Loader, error) {
	root, f, err := fSys.CleanedAbs(remote.AbsPath())
	if err != nil {
		_ = cleaner()
		return nil, errors.Wrap(err)
	}
	// We don't know that the path requested in the remote
	// is a directory until we actually fetch it and look
	// inside.  That just happened, hence the error check
	// is here.
	if f != "" {
		_ = cleaner()
		return nil, fmt.Errorf(
			"'%s' refers to file '%s'; expecting directory",
			remote.AbsPath(), f)
	}
	// Path in the remote can contain symlinks that exit it. We can
	// only check for this after fetching it.
	if !root.HasPrefix(remote.FetchDir()) {
		_ = cleaner()
		_, unit := remoteKind(remote)
		return nil, fmt.Errorf("%q refers to directory outside of %s %q", remote.AbsPath(),
			unit, remote.FetchDir())
	}
	return &FileLoader{
		// Clones and pulls never allowed to escape root.
		loadRestrictor: RestrictionRootOnly,
		root:           root,
		referrer:       referrer,
		remote:         remote,
		fSys:           fSys,
		cloner:         cloner,
		puller:         puller,
		remotes:        remotes,
		cleaner:        cleaner,
	}, nil
}

func (fl *FileLoader) errIfContainmentViolation(
	base filesys.ConfirmedDir) error {
	containingRepo := fl.containingRepo()
	if containingRepo == nil {
		return nil
	}
	if !base.HasPrefix(containingRepo.FetchDir()) {
		fetched, unit := remoteKind(containingRepo)
		return fmt.Errorf(
			"security; bases in kustomizations found in "+
				"%s must be within the %s, "+
				"but base '%s' is outside '%s'",
			fetched, unit, base, containingRepo.FetchDir())
	}
	return nil
}

// Looks back through referrers for the nearest remote root,
// a git repo or an OCI artifact, returning nil if none found.
func (fl *FileLoader) containingRepo() remoteSpec {
	if fl.remote != nil {
		return fl.remote
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.containingRepo()
}

// errIfArgEqualOrHigher tests whether the argument,
// is equal to or above the root of any ancestor,
// up to the nearest remote root.
func (fl *FileLoader) errIfArgEqualOrHigher(
	candidateRoot filesys.ConfirmedDir) error {
	if fl.root.HasPrefix(candidateRoot) {
		return fmt.Errorf(
			"cycle detected: candidate root '%s' contains visited root '%s'",
			candidateRoot, fl.root)
	}
	// The roots above the nearest remote root are outside of it, or in
	// it only because the build refers to the same remote again, which
	// shares the clone or pull; cycles through remotes are detected by
	// errIfRepoCycle.
	if fl.referrer == nil || fl.remote != nil {
		return nil
	}
	return fl.referrer.errIfArgEqualOrHigher(candidateRoot)
}

// TODO(monopole): Distinguish branches?
// I.e. Allow a distinction between git URI with
// path foo and tag bar and a git URI with the same
// path but a different tag?
func (fl *FileLoader) errIfRepoCycle(newRepoSpec remoteSpec) error {
	// TODO(monopole): Use parsed data instead of Raw().
	if fl.remote != nil &&
		strings.HasPrefix(fl.remote.Raw(), newRepoSpec.Raw()) {
		return fmt.Errorf(
			"cycle detected: URI '%s' referenced by previous URI '%s'",
			newRepoSpec.Raw(), fl.remote.Raw())
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.errIfRepoCycle(newRepoSpec)
}

// Load returns the content of file at the given path,
// else an error. Relative paths are taken relative
// to the root.
func (fl *FileLoader) Load(path string) ([]byte, error) {
	if IsRemoteFile(path) {
		return fl.httpClientGetContent(path)
	}
	if !filepath.IsAbs(path) {
		path = fl.root.Join(path)
	}
	path, err := fl.loadRestrictor(fl.fSys, fl.root, path)
	if err != nil {
		return nil, err
	}
	content, err := fl.fSys.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return content, nil
}

func (fl *FileLoader) httpClientGetContent(path string) ([]byte, error) {
	var hc *http.Client
	if fl.http != nil {
		hc = fl.http
	} else {
		hc = &http.Client{}
	}
	parsedURL, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	resp, err := hc.Get(parsedURL.String())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer resp.Body.Close()
	// response unsuccessful
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, err = git.NewRepoSpecFromURL(path)
		if err == nil {
			return nil, errors.Errorf("URL is a git repository")
		}
		return nil, fmt.Errorf("%w: status code %d (%s)", ErrHTTP, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	content, err := io.ReadAll(resp.Body)
	return content, errors.Wrap(err)
}

// Cleanup runs the cleaner, or, for the loader at the top of a
// build (it has no referrer), removes the clones and pulls of the
// build, its own included.
func (fl *FileLoader) Cleanup() error {
	if fl.referrer == nil && fl.remotes != nil {
		return fl.remotes.cleanup()
	}
	return fl.cleaner()
}
