// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/mitchellh/go-homedir"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// IsRemoteFile returns whether path has a url scheme that kustomize allows for
// remote files. See https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md
func IsRemoteFile(path string) (bool, *url.URL) {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https"), u
}

// fileLoader is a kustomization's interface to files.
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
//   `Load` is used to visit these paths.
//
//   These paths refer to resources, patches,
//   data for ConfigMaps and Secrets, etc.
//
//   The loadRestrictor may disallow certain paths
//   or classes of paths.
//
// * bases (other kustomizations)
//
//   `New` is used to load bases.
//
//   A base can be either a remote git repo URL, or
//   a directory specified relative to the current
//   root. In the former case, the repo is locally
//   cloned, and the new loader is rooted on a path
//   in that clone.
//
//   As loaders create new loaders, a root history
//   is established, and used to disallow:
//
//   - A base that is a repository that, in turn,
//     specifies a base repository seen previously
//     in the loading stack (a cycle).
//
//   - An overlay depending on a base positioned at
//     or above it.  I.e. '../foo' is OK, but '.',
//     '..', '../..', etc. are disallowed.  Allowing
//     such a base has no advantages and encourages
//     cycles, particularly if some future change
//     were to introduce globbing to file
//     specifications in the kustomization file.
//
// These restrictions assure that kustomizations
// are self-contained and relocatable, and impose
// some safety when relying on remote kustomizations,
// e.g. a remotely loaded ConfigMap generator specified
// to read from /etc/passwd will fail.
//
type fileLoader struct {
	// Loader that spawned this loader.
	// Used to avoid cycles.
	referrer *fileLoader

	// An absolute, cleaned path to a directory.
	// The Load function will read non-absolute
	// paths relative to this directory.
	root filesys.ConfirmedDir

	// Restricts behavior of Load function.
	loadRestrictor LoadRestrictorFunc

	// If this is non-nil, the files were
	// obtained from the given repository.
	repoSpec *git.RepoSpec

	// File system utilities.
	fSys filesys.FileSystem

	// Used to load from HTTP
	http *http.Client

	// Used to clone repositories.
	cloner git.Cloner

	// Used to clean up, as needed.
	cleaner func() error
}

// NewFileLoaderAtCwd returns a loader that loads from PWD.
// A convenience for kustomize edit commands.
func NewFileLoaderAtCwd(fSys filesys.FileSystem) *fileLoader {
	return newLoaderOrDie(
		RestrictionRootOnly, fSys, filesys.SelfDir)
}

// NewFileLoaderAtRoot returns a loader that loads from "/".
// A convenience for tests.
func NewFileLoaderAtRoot(fSys filesys.FileSystem) *fileLoader {
	return newLoaderOrDie(
		RestrictionRootOnly, fSys, filesys.Separator)
}

// Repo returns the absolute path to the repo that contains Root if this fileLoader was created from a url
// or the empty string otherwise.
func (fl *fileLoader) Repo() string {
	if fl.repoSpec != nil {
		return fl.repoSpec.Dir.String()
	}
	return ""
}

// Root returns the absolute path that is prepended to any
// relative paths used in Load.
func (fl *fileLoader) Root() string {
	return fl.root.String()
}

func newLoaderOrDie(
	lr LoadRestrictorFunc,
	fSys filesys.FileSystem, path string) *fileLoader {
	root, err := filesys.ConfirmDir(fSys, path)
	if err != nil {
		log.Fatalf("unable to make loader at '%s'; %v", path, err)
	}
	return newLoaderAtConfirmedDir(
		lr, root, fSys, nil, git.ClonerUsingGitExec)
}

// newLoaderAtConfirmedDir returns a new fileLoader with given root.
func newLoaderAtConfirmedDir(
	lr LoadRestrictorFunc,
	root filesys.ConfirmedDir, fSys filesys.FileSystem,
	referrer *fileLoader, cloner git.Cloner) *fileLoader {
	return &fileLoader{
		loadRestrictor: lr,
		root:           root,
		referrer:       referrer,
		fSys:           fSys,
		cloner:         cloner,
		cleaner:        func() error { return nil },
	}
}

// New returns a new Loader, rooted relative to current loader,
// or rooted in a temp directory holding a git repo clone.
func (fl *fileLoader) New(path string) (ifc.Loader, error) {
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
			repoSpec, fl.fSys, fl, fl.cloner)
	}

	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("new root '%s' cannot be absolute", path)
	}
	root, err := filesys.ConfirmDir(fl.fSys, fl.root.Join(path))
	if err != nil {
		return nil, errors.WrapPrefixf(err, ErrRtNotDir.Error())
	}
	if err = fl.errIfGitContainmentViolation(root); err != nil {
		return nil, err
	}
	if err = fl.errIfArgEqualOrHigher(root); err != nil {
		return nil, err
	}
	return newLoaderAtConfirmedDir(
		fl.loadRestrictor, root, fl.fSys, fl, fl.cloner), nil
}

// newLoaderAtGitClone returns a new Loader pinned to a temporary
// directory holding a cloned git repo.
func newLoaderAtGitClone(
	repoSpec *git.RepoSpec, fSys filesys.FileSystem,
	referrer *fileLoader, cloner git.Cloner) (ifc.Loader, error) {
	cleaner := repoSpec.Cleaner(fSys)
	err := cloner(repoSpec)
	if err != nil {
		cleaner()
		return nil, err
	}
	root, f, err := fSys.CleanedAbs(repoSpec.AbsPath())
	if err != nil {
		cleaner()
		return nil, err
	}
	// We don't know that the path requested in repoSpec
	// is a directory until we actually clone it and look
	// inside.  That just happened, hence the error check
	// is here.
	if f != "" {
		cleaner()
		return nil, fmt.Errorf(
			"'%s' refers to file '%s'; expecting directory",
			repoSpec.AbsPath(), f)
	}
	// Path in repo can contain symlinks that exit repo. We can only
	// check for this after cloning repo.
	if !root.HasPrefix(repoSpec.CloneDir()) {
		_ = cleaner()
		return nil, fmt.Errorf("%q refers to directory outside of repo %q", repoSpec.AbsPath(),
			repoSpec.CloneDir())
	}
	return &fileLoader{
		// Clones never allowed to escape root.
		loadRestrictor: RestrictionRootOnly,
		root:           root,
		referrer:       referrer,
		repoSpec:       repoSpec,
		fSys:           fSys,
		cloner:         cloner,
		cleaner:        cleaner,
	}, nil
}

func (fl *fileLoader) errIfGitContainmentViolation(
	base filesys.ConfirmedDir) error {
	containingRepo := fl.containingRepo()
	if containingRepo == nil {
		return nil
	}
	if !base.HasPrefix(containingRepo.CloneDir()) {
		return fmt.Errorf(
			"security; bases in kustomizations found in "+
				"cloned git repos must be within the repo, "+
				"but base '%s' is outside '%s'",
			base, containingRepo.CloneDir())
	}
	return nil
}

// Looks back through referrers for a git repo, returning nil
// if none found.
func (fl *fileLoader) containingRepo() *git.RepoSpec {
	if fl.repoSpec != nil {
		return fl.repoSpec
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.containingRepo()
}

// errIfArgEqualOrHigher tests whether the argument,
// is equal to or above the root of any ancestor.
func (fl *fileLoader) errIfArgEqualOrHigher(
	candidateRoot filesys.ConfirmedDir) error {
	if fl.root.HasPrefix(candidateRoot) {
		return fmt.Errorf(
			"cycle detected: candidate root '%s' contains visited root '%s'",
			candidateRoot, fl.root)
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.errIfArgEqualOrHigher(candidateRoot)
}

// TODO(monopole): Distinguish branches?
// I.e. Allow a distinction between git URI with
// path foo and tag bar and a git URI with the same
// path but a different tag?
func (fl *fileLoader) errIfRepoCycle(newRepoSpec *git.RepoSpec) error {
	// TODO(monopole): Use parsed data instead of Raw().
	if fl.repoSpec != nil &&
		strings.HasPrefix(fl.repoSpec.Raw(), newRepoSpec.Raw()) {
		return fmt.Errorf(
			"cycle detected: URI '%s' referenced by previous URI '%s'",
			newRepoSpec.Raw(), fl.repoSpec.Raw())
	}
	if fl.referrer == nil {
		return nil
	}
	return fl.referrer.errIfRepoCycle(newRepoSpec)
}

// Load returns the content of file at the given path,
// else an error. Relative paths are taken relative
// to the root.
func (fl *fileLoader) Load(path string) ([]byte, error) {
	if remote, u := IsRemoteFile(path); remote {
		return fl.httpClientGetContent(u)
	}
	if !filepath.IsAbs(path) {
		path = fl.root.Join(path)
	}
	path, err := fl.loadRestrictor(fl.fSys, fl.root, path)
	if err != nil {
		return nil, err
	}
	return fl.fSys.ReadFile(path)
}

func (fl *fileLoader) httpClientGetContent(u *url.URL) ([]byte, error) {
	var hc *http.Client
	if fl.http != nil {
		hc = fl.http
	} else {
		hc = &http.Client{}
	}
	err := addAuthFromNetrc(u)
	if err != nil {
		return nil, err
	}
	resp, err := hc.Get(u.String())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer resp.Body.Close()
	// response unsuccessful
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, err = git.NewRepoSpecFromURL(u.String())
		if err == nil {
			return nil, errors.Errorf("URL is a git repository")
		}
		return nil, fmt.Errorf("%w: status code %d (%s)", ErrHTTP, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	content, err := io.ReadAll(resp.Body)
	return content, errors.Wrap(err)
}

// Cleanup runs the cleaner.
func (fl *fileLoader) Cleanup() error {
	return fl.cleaner()
}

// addAuthFromNetrc adds auth information to the URL from the user's
// netrc file if it can be found. This will only add the auth info
// if the URL doesn't already have auth info specified and the
// the username is blank.
func addAuthFromNetrc(u *url.URL) error {
	// If the URL already has auth information, do nothing
	if u.User != nil && u.User.Username() != "" {
		return nil
	}

	// Get the netrc file path
	path := os.Getenv("NETRC")
	if path == "" {
		filename := ".netrc"
		if runtime.GOOS == "windows" {
			filename = "_netrc"
		}

		var err error
		path, err = homedir.Expand("~/" + filename)
		if err != nil {
			return err
		}
	}

	// If the file is not a file, then do nothing
	if fi, err := os.Stat(path); err != nil {
		// File doesn't exist, do nothing
		if os.IsNotExist(err) {
			return nil
		}

		// Some other error!
		return err
	} else if fi.IsDir() {
		// File is directory, ignore
		return nil
	}

	// Load up the netrc file
	netrc, err := netrc.ParseFile(path)
	if err != nil {
		return fmt.Errorf("Error parsing netrc file at %q: %s", path, err)
	}

	machine := netrc.FindMachine(u.Host)
	if machine == nil {
		// Machine not found, no problem
		return nil
	}

	// Set the user info
	u.User = url.UserPassword(machine.Login, machine.Password)
	return nil
}
