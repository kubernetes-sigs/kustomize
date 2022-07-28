// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const urlSeparator = "/"
const replacement = "-"
const dstPrefix = "localized"
const fatalPrefix = "should never happen"
const errRemoteTarget = "invalid remote localize target '%s'"

var errRootOutsideScope = errors.Errorf("local root not within scope")

// locArgs stores localize arguments
type locArgs struct {

	// directory that bounds target's local references, empty string if target is remote
	scope filesys.ConfirmedDir

	// localize destination
	newDir string
}

// locLoader is the Loader for kustomize localize.
//
// Each locLoader is created and Loads under localize constraints. For each input path, locLoader
// also outputs the corresponding path to the localized copy.
type locLoader struct {
	fSys filesys.FileSystem

	args *locArgs

	// loader at locLoader's current directory
	ldr ifc.Loader

	// whether locLoader and all its ancestors are the result of local references
	local bool

	// directory in newDir that mirrors ldr root
	dst string
}

// validateScope returns the scope given localize arguments and targetLdr that corresponds to targetArg
func validateScope(scopeArg string, targetArg string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	var scope filesys.ConfirmedDir

	errBadScopeForTarget := fmt.Sprintf("invalid scope '%s' for target '%s'", scopeArg, targetArg)
	switch _, remote := targetLdr.Repo(); {
	case remote:
		if scopeArg != "" {
			return "", errors.WrapPrefixf(
				errors.Errorf("cannot specify scope for remote target"),
				errBadScopeForTarget)
		}
	case scopeArg == "":
		scope = filesys.ConfirmedDir(targetLdr.Root())
	default:
		var err error
		scope, err = filesys.ConfirmDir(fSys, scopeArg)
		if err != nil {
			return "", errors.WrapPrefixf(err, "unable to establish scope '%s'", scopeArg)
		}
		if !filesys.ConfirmedDir(targetLdr.Root()).HasPrefix(scope) {
			return scope, errors.WrapPrefixf(errRootOutsideScope, errBadScopeForTarget)
		}
	}
	return scope, nil
}

// parseLocRootURL accepts a remote root rootURL and returns its RepoSpec if it satisfies localize requirements;
// otherwise, an error.
func parseLocRootURL(rootURL string) (*git.RepoSpec, error) {
	spec, err := git.NewRepoSpecFromURL(rootURL)
	if err != nil {
		log.Fatalf(errors.WrapPrefixf(
			err,
			"%s: cannot parse validated remote root '%s'", fatalPrefix, rootURL).Error())
	}
	if spec.Ref == "" {
		return nil, errors.Errorf("localize remote root missing ref query string parameter")
	}
	return spec, nil
}

// urlBase is the url equivalent of filepath.Base
func urlBase(url string) string {
	cleaned := strings.TrimRight(url, urlSeparator)
	i := strings.LastIndex(cleaned, urlSeparator)
	if i < 0 {
		return cleaned
	}
	return cleaned[i+1:]
}

// defaultNewDir calculates the path of the default localize destination directory from targetLdr
// at the localize target and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (string, error) {
	wd, err := filesys.ConfirmDir(fSys, filesys.SelfDir)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to establish working directory")
	}

	var newDir string
	targetDir := filepath.Base(targetLdr.Root())
	if repo, remote := targetLdr.Repo(); remote {
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		newDir = strings.Join(
			[]string{dstPrefix, targetDir, strings.ReplaceAll(spec.Ref, urlSeparator, replacement)},
			replacement)
	} else {
		newDir = strings.Join([]string{dstPrefix, targetDir}, replacement)
	}
	return wd.Join(newDir), nil
}

// NewLocLoader is the factory method for locLoader.
// NewLocLoader accepts the arguments to kustomize localize and returns a locLoader at targetArg.
func NewLocLoader(targetArg string, scopeArg string, newDirArg string, fSys filesys.FileSystem) (*locLoader, error) {
	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, targetArg, fSys)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "unable to establish localize target '%s'", targetArg)
	}
	var spec *git.RepoSpec
	if _, remote := ldr.Repo(); remote {
		spec, err = parseLocRootURL(targetArg)
		if err != nil {
			return nil, errors.WrapPrefixf(err, errRemoteTarget, targetArg)
		}
	}

	scope, err := validateScope(scopeArg, targetArg, ldr, fSys)
	if err != nil {
		clean(ldr)
		return nil, err
	}

	var newDir string
	if newDirArg == "" {
		newDir, err = defaultNewDir(ldr, spec, fSys)
		if err != nil {
			clean(ldr)
			return nil, err
		}
	} else {
		newDir = newDirArg
	}

	var realScope string
	if repo, remote := ldr.Repo(); remote {
		realScope = repo
	} else {
		realScope = scope.String()
	}
	toDst, err := filepath.Rel(realScope, ldr.Root())
	if err != nil {
		log.Fatalf(errors.WrapPrefixf(
			err,
			"%s: scope '%s' contains target '%s'", fatalPrefix, realScope, ldr.Root()).Error())
	}

	return &locLoader{
		fSys: fSys,
		args: &locArgs{
			scope:  scope,
			newDir: newDir,
		},
		ldr:   ldr,
		local: scope == "",
		dst:   filepath.Join(newDir, toDst),
	}, nil
}

// Cleanup tries to clean the by-products of localize and logs if unsuccessful
func (ll *locLoader) Cleanup() {
	clean(ll.ldr)
}

func clean(ldr ifc.Loader) {
	if err := ldr.Cleanup(); err != nil {
		log.Printf("%s", errors.WrapPrefixf(err, "unable to clean by-products of loading target").Error())
	}
}
