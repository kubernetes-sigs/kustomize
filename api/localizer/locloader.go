// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const urlSeparator = "/"
const localizeDir = "localized-files"

const errAlpha = "alpha unsupported feature"
const errMkdir = "unable to create directory in localize destination"
const errRef = "invalid reference '%s' at '%s'"

var errLocDirExists = errors.Errorf(localizeDir + " already exists")
var errDstRef = errors.Errorf("reference into localize destination")
var errRemoteRootNoRef = errors.Errorf("localize remote repo missing ref")

type args struct {
	target filesys.ConfirmedDir

	scope filesys.ConfirmedDir

	newDir filesys.ConfirmedDir
}

// locLoader, associated with a directory, is the file loader for kustomize localize
type locLoader struct {
	args *args

	fSys filesys.FileSystem

	ldr ifc.Loader

	local bool

	// directory in newDir that mirrors root
	dst filesys.ConfirmedDir
}

// TODO: decide what to do with this, or replace with Rel
func suffix(dir string, inDir string) string {
	i := len(dir)
	if len(inDir) > i && inDir[len(dir)] == filepath.Separator {
		i = len(dir) + 1
	}
	return inDir[i:]
}

// TODO: handle default value for optional args

// newLocLoader is the factory method for locLoader.
// newLocLoader accepts the arguments to kustomize localize and returns a locLoader at target.
func newLocLoader(targetArg string, scopeArg string, newDirArg string, fSys filesys.FileSystem) (*locLoader, error) {
	// handle newDir before target, in case target download creates newDir by accident
	if fSys.Exists(newDirArg) {
		return nil, errors.WrapPrefixf(
			errors.Errorf("localize destination already exists"),
			fmt.Sprintf("invalid localize destination '%s'", newDirArg))
	}
	if err := fSys.MkdirAll(newDirArg); err != nil {
		return nil, errors.WrapPrefixf(err, "unable to create localize destination")
	}
	// TODO: add newDir removal for errors
	newDir, err := filesys.ConfirmDir(fSys, newDirArg)
	if err != nil {
		return nil, errors.WrapPrefixf(err, fmt.Sprintf("unable to establish localize destination"))
	}

	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, targetArg, fSys)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to establish localize target")
	}

	var scope filesys.ConfirmedDir
	var inNewDir string
	if repo, remote := ldr.Repo(); remote {
		if scopeArg != "" {
			ldr.Cleanup()
			return nil, errors.WrapPrefixf(
				errors.Errorf("repo is scope for remote target"),
				fmt.Sprintf("should not specify scope ('%s') for remote localize target '%s'", scopeArg, targetArg))
		}
		spec, err := git.NewRepoSpecFromURL(targetArg)
		if err != nil {
			panic(err)
		}
		if spec.Ref == "" {
			ldr.Cleanup()
			return nil, errors.WrapPrefixf(errRemoteRootNoRef, "invalid localize target '%s'", targetArg)
		}
		inNewDir = suffix(repo, ldr.Root())
	} else { // local target
		if scopeArg == "" {
			scope = filesys.ConfirmedDir(ldr.Root())
			inNewDir = filesys.SelfDir
		} else {
			if scope, err = filesys.ConfirmDir(fSys, scopeArg); err != nil {
				return nil, errors.WrapPrefixf(err, "unable to establish localize scope")
			}
			if !filesys.ConfirmedDir(ldr.Root()).HasPrefix(scope) {
				return nil, errors.WrapPrefixf(
					errors.Errorf("localize scope does not contain localize target"),
					fmt.Sprintf("invalid localize scope '%s' for localize target '%s'", scopeArg, targetArg))
			}
			inNewDir = suffix(scope.String(), ldr.Root())
		}
	}

	if err = fSys.MkdirAll(newDir.Join(inNewDir)); err != nil {
		ldr.Cleanup()
		return nil, errors.WrapPrefixf(err, errMkdir)
	}

	return &locLoader{
		args: &args{
			target: filesys.ConfirmedDir(ldr.Root()),
			scope:  scope,
			newDir: newDir,
		},
		fSys:  fSys,
		local: scopeArg != "",
		ldr:   ldr,
		dst:   filesys.ConfirmedDir(newDir.Join(inNewDir)),
	}, nil
}

func (ll *locLoader) dest() filesys.ConfirmedDir {
	return ll.dst
}

func toPath(urlPath string) string {
	return filepath.Join(strings.Split(urlPath, urlSeparator)...)
}

func locFilePath(u *url.URL) string {
	// raw github urls are only type of file urls kustomize accepts;
	// path consists of org, repo, version, path in repo
	return filepath.Join(localizeDir, u.Hostname(), toPath(u.Path))
}

// load returns the contents of path and its localize destination if path is a file that can be localized;
// otherwise, load returns an error
func (ll *locLoader) load(path string) ([]byte, string, error) {
	// checks in root, and thus in scope
	content, err := ll.ldr.Load(path)
	if err != nil {
		return nil, "", errors.WrapPrefixf(err, "unable to read localize target reference file")
	}

	var fDst string
	switch {
	case loader.HasRemoteFileScheme(path):
		// TODO: fix paths displayed in errors to be reader-comprehensible
		if ll.fSys.Exists(filepath.Join(ll.ldr.Root(), localizeDir)) {
			return nil, "", errors.WrapPrefixf(errLocDirExists, "at '%s'", ll.ldr.Root())
		}

		u, err := url.Parse(path)
		if err != nil {
			panic(err)
		}
		fDst = locFilePath(u)
	case !filepath.IsAbs(path):
		// avoid symlinks; only write file corresponding to actual location in root
		// avoid path that Load() shows to be in root, but may traverse outside
		// temporarily; for example, ../root/config; problematic for rename and
		// relocation
		dir, f, err := ll.fSys.CleanedAbs(filepath.Join(ll.ldr.Root(), path))
		if err != nil {
			panic(err)
		}
		// if root originally url, newDir can contain copy;
		// if newDir in scope or target, path can reference newDir without leaving root
		if !filesys.ConfirmedDir(ll.ldr.Root()).HasPrefix(ll.args.newDir) && dir.HasPrefix(ll.args.newDir) {
			return nil, "", errors.WrapPrefixf(errDstRef, errRef, path, ll.ldr.Root())
		}
		fDst = suffix(ll.ldr.Root(), dir.Join(f))
	case filepath.IsAbs(path):
		// TODO: fix error wrapping
		return nil, "", errors.WrapPrefixf(
			errors.WrapPrefixf(errors.Errorf("absolute path"), errRef, path, ll.ldr.Root()),
			errAlpha)
	default:
		return nil, "", errors.WrapPrefixf("unrecognized case", errRef, path, ll.ldr.Root())
	}

	return content, fDst, nil
}

func (ll *locLoader) new(path string) (*locLoader, string, error) {
	ldr, err := ll.ldr.New(path)
	if err != nil {
		return nil, "", err
	}

	var inDst string
	if repo, remote := ldr.Repo(); remote {
		// TODO: add Hostname() to repospec?
		u, err := url.Parse(path)
		if err != nil {
			panic(err)
		}
		spec, err := git.NewRepoSpecFromURL(path)
		if err != nil {
			panic(err)
		}
		if spec.Ref == "" {
			ldr.Cleanup()
			return nil, "", errors.WrapPrefixf(errRemoteRootNoRef, errRef, path, ll.ldr.Root())
		}
		if ll.fSys.Exists(filepath.Join(ll.ldr.Root(), localizeDir)) {
			ldr.Cleanup()
			return nil, "", errors.WrapPrefixf(errLocDirExists, "at '%s'", ll.ldr.Root())
		}
		inRepo := suffix(repo, ldr.Root())
		inDst = filepath.Join(localizeDir, u.Hostname(), toPath(spec.OrgRepo), toPath(spec.Ref), inRepo)
	} else {
		if ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.scope) {
			// TODO: fix error msg construction
			return nil, "", errors.WrapPrefixf(
				errors.Errorf("referenced local root outside localize scope"),
				errRef+" for localize scope '%s'", path, ll.ldr.Root(), ll.args.scope)
		}
		if !filesys.ConfirmedDir(ll.ldr.Root()).HasPrefix(ll.args.newDir) &&
			filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.newDir) {
			return nil, "", errors.WrapPrefixf(errDstRef, errRef, path, ll.ldr.Root())
		}
		inDst, err = filepath.Rel(ll.ldr.Root(), ldr.Root())
		if err != nil {
			panic(err)
		}
	}
	if err = ll.fSys.MkdirAll(ll.dst.Join(inDst)); err != nil {
		ldr.Cleanup()
		return nil, "", errors.WrapPrefixf(err, errMkdir)
	}
	// make destination dir
	return &locLoader{
		args:  ll.args,
		fSys:  ll.fSys,
		ldr:   ldr,
		local: ll.local && strings.HasPrefix(inDst, localizeDir),
		dst:   filesys.ConfirmedDir(ll.dst.Join(inDst)),
	}, inDst, nil
}

// TODO: this should be internal; external cleanup should only clean ldr
func (ll *locLoader) cleanup() error {
	err := ll.ldr.Cleanup()
	if err != nil {
		return errors.WrapPrefixf(err, "unable to clean up localize target reference")
	}

	var dstDir filesys.ConfirmedDir
	if ll.args.target.String() == ll.ldr.Root() {
		dstDir = ll.args.newDir
	} else {
		dstDir = ll.dst
	}
	err = ll.fSys.RemoveAll(dstDir.String())
	if err != nil {
		return errors.WrapPrefixf(err, "unable to clean up directory in localize destination")
	}

	return nil
}
