// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"fmt"
	"log"
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
const replacement = "-"

const dstPrefix = "localized"
const localizeDir = dstPrefix + replacement + "files"
const fatalPrefix = "should never happen"
const errAlpha = "alpha unsupported localize feature"
const errReference = "invalid localize reference location '%s'"

var errDstRef = errors.Errorf("reference into localize destination")
var errRootOutsideScope = errors.Errorf("local root not within localize scope")

// locArgs stores localize arguments
type locArgs struct {

	// directory that bounds target's local references, empty string if target is remote
	scope filesys.ConfirmedDir

	// localize destination
	newDir filesys.ConfirmedDir
}

// LocLoader is the Loader for kustomize localize.
//
// Each LocLoader is created and Loads under localize constraints. For each input path, LocLoader
// also outputs the corresponding path to the localized copy.
type LocLoader struct {
	fSys filesys.FileSystem

	args *locArgs

	// loader at LocLoader's current directory
	ldr ifc.Loader

	// whether LocLoader and all its ancestors are the result of local references
	local bool

	// directory in newDir that mirrors ldr root
	dst string
}

// LocPath represents the localized counterpart of a target reference
type LocPath struct {
	Remote bool

	Path string
}

// ValidateLocArgs validates the arguments to kustomize localize.
// On success, ValidateLocArgs returns a LocLoader at targetArg and the localize destination directory; otherwise, an
// error.
func ValidateLocArgs(targetArg string, scopeArg string, newDirArg string, fSys filesys.FileSystem) (*LocLoader, filesys.ConfirmedDir, error) {
	// for security, should enforce load restrictions
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, targetArg, fSys)
	if err != nil {
		return nil, "", errors.WrapPrefixf(err, "unable to establish localize target '%s'", targetArg)
	}
	var spec *git.RepoSpec
	if _, remote := ldr.Repo(); remote {
		spec, err = parseLocRootURL(targetArg)
		if err != nil {
			cleanLdr(ldr)
			return nil, "", errors.WrapPrefixf(err, "invalid remote localize target '%s'", targetArg)
		}
	}

	scope, err := validateScope(scopeArg, targetArg, ldr, fSys)
	if err != nil {
		cleanLdr(ldr)
		return nil, "", errors.WrapPrefixf(err, "invalid localize scope '%s'", scopeArg)
	}

	newDir, err := validateNewDir(newDirArg, ldr, spec, fSys)
	if err != nil {
		cleanLdr(ldr)
		return nil, "", errors.WrapPrefixf(err, "invalid localize destination '%s'", newDirArg)
	}

	var realScope string
	if repo, remote := ldr.Repo(); remote {
		realScope = repo
	} else {
		realScope = scope.String()
	}
	toDst := pathSuffix(realScope, ldr.Root())

	return &LocLoader{
		fSys: fSys,
		args: &locArgs{
			scope:  scope,
			newDir: newDir,
		},
		ldr:   ldr,
		local: scope != "",
		dst:   newDir.Join(toDst),
	}, newDir, nil
}

// validateScope returns the scope given localize arguments and targetLdr that corresponds to targetArg
func validateScope(scopeArg string, targetArg string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	var scope filesys.ConfirmedDir

	errBadScopeForTarget := fmt.Sprintf("invalid localize scope '%s' for target '%s'", scopeArg, targetArg)
	switch _, remote := targetLdr.Repo(); {
	case remote:
		if scopeArg != "" {
			return "", errors.WrapPrefixf(
				errors.Errorf("cannot specify scope for remote localize target"),
				errBadScopeForTarget)
		}
	case scopeArg == "":
		scope = filesys.ConfirmedDir(targetLdr.Root())
	default:
		var err error
		scope, err = filesys.ConfirmDir(fSys, scopeArg)
		if err != nil {
			return "", errors.WrapPrefixf(err, "unable to establish localize scope")
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
		return nil, errors.WrapPrefixf(
			errors.Errorf("localize remote root missing ref query string parameter"),
			"invalid localize remote root '%s'", rootURL)
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

// defaultNewDir calculates the default localize destination directory name from targetLdr at the localize target
// and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec) string {
	targetDir := filepath.Base(targetLdr.Root())
	switch repo, remote := targetLdr.Repo(); {
	case remote:
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		return strings.Join(
			[]string{dstPrefix, targetDir, strings.ReplaceAll(spec.Ref, urlSeparator, replacement)},
			replacement)
	case targetDir == string(filepath.Separator):
		return dstPrefix
	default:
		return strings.Join([]string{dstPrefix, targetDir}, replacement)
	}
}

// validateNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func validateNewDir(newDirArg string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	var newDirPath string
	if newDirArg == "" {
		newDirPath = defaultNewDir(targetLdr, spec)
	} else {
		newDirPath = newDirArg
	}
	if fSys.Exists(newDirPath) {
		return "", errors.WrapPrefixf(
			errors.Errorf("localize destination already exists"),
			"invalid localize destination path '%s'", newDirPath)
	}
	if err := fSys.Mkdir(newDirPath); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, newDirPath)
	if err != nil {
		cleanDst(newDirPath, fSys)
		return "", errors.WrapPrefixf(err, "unable to establish localize destination")
	}

	return newDir, nil
}

// pathSuffix returns path relative to prefix, where both are cleaned absolute paths
// and prefix is a directory known to contain path
func pathSuffix(prefix string, path string) string {
	if prefix == string(filepath.Separator) || prefix == path {
		return path[len(prefix):]
	}
	return path[len(prefix)+1:]
}

func (ll *LocLoader) Root() string {
	return ll.ldr.Root()
}

func (ll *LocLoader) Dst() string {
	return ll.dst
}

func toFSPath(urlPath string) string {
	return filepath.Join(strings.Split(urlPath, urlSeparator)...)
}

func locFilePath(u *url.URL) string {
	// raw github urls are the only type of file urls kustomize accepts;
	// path consists of org, repo, version, path in repo
	return filepath.Join(localizeDir, u.Hostname(), toFSPath(u.Path))
}

// Load returns the contents of path and its corresponding localize destination path information
// if path is a file that can be localized. Otherwise, Load returns an error.
func (ll *LocLoader) Load(path string) ([]byte, *LocPath, error) {
	// checks in root, and thus in scope
	content, err := ll.ldr.Load(path)
	if err != nil {
		return nil, nil, errors.WrapPrefixf(err, "invalid file reference")
	}

	var dst string
	var remote bool
	switch {
	case loader.HasRemoteFileScheme(path):
		remote = true

		u, err := url.Parse(path)
		if err != nil {
			log.Fatalf(errors.WrapPrefixf(
				err,
				"%s: cannot parse validated file url '%s'", fatalPrefix, path).Error())
		}
		dst = locFilePath(u)
	case !filepath.IsAbs(path):
		abs := filepath.Join(ll.ldr.Root(), path)
		// avoid symlinks; only write file corresponding to actual location in root
		// avoid path that Load() shows to be in root, but may traverse outside
		// temporarily; for example, ../root/config; problematic for rename and
		// relocation
		dir, f, err := ll.fSys.CleanedAbs(abs)
		if err != nil {
			log.Fatalf(errors.WrapPrefixf(
				err,
				"%s: cannot clean validated file path '%s'", fatalPrefix, abs).Error())
		}

		cleanPath := dir.Join(f)
		// target cannot reference newDir, as this load would've failed prior to localize;
		// not a problem if remote because then reference could only be in newDir if repo copy,
		// which will be cleaned, is inside newDir
		if ll.local && dir.HasPrefix(ll.args.newDir) {
			return nil, nil, errors.WrapPrefixf(errDstRef, errReference, cleanPath)
		}
		dst = pathSuffix(ll.ldr.Root(), cleanPath)
	default:
		return nil, nil, errors.WrapPrefixf(errors.Errorf(errAlpha), "path '%s' is absolute", path)
	}

	return content, &LocPath{
		remote,
		dst,
	}, nil
}

// New returns a LocLoader at path and whether path is remote if path is a root that can be localized.
// Otherwise, New returns an error.
func (ll *LocLoader) New(path string) (*LocLoader, *LocPath, error) {
	ldr, err := ll.ldr.New(path)
	if err != nil {
		return nil, nil, errors.WrapPrefixf(err, "invalid root reference")
	}

	var repo, inDst string
	var remote bool
	switch repo, remote = ldr.Repo(); {
	case remote:
		spec, err := parseLocRootURL(path)
		if err != nil {
			return nil, nil, err
		}
		inRepo := pathSuffix(repo, ldr.Root())
		inDst = filepath.Join(localizeDir, spec.Domain, toFSPath(spec.OrgRepo), toFSPath(spec.Ref), inRepo)
	case ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.scope):
		return nil, nil, errors.WrapPrefixf(
			errRootOutsideScope,
			"invalid root location '%s' for scope '%s'", ldr.Root(), ll.args.scope)
	case ll.local && filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.newDir):
		return nil, nil, errors.WrapPrefixf(errDstRef, errReference, ldr.Root())
	default:
		inDst, err = filepath.Rel(ll.ldr.Root(), ldr.Root())
		if err != nil {
			log.Fatalf(errors.WrapPrefixf(err,
				"%s: cannot find relative path from root '%s' to its valid root reference '%s'",
				fatalPrefix, ll.ldr.Root(), ldr.Root()).Error())
		}
	}

	return &LocLoader{
			fSys:  ll.fSys,
			args:  ll.args,
			ldr:   ldr,
			local: ll.local && !remote,
			dst:   filepath.Join(ll.dst, inDst),
		}, &LocPath{
			remote,
			inDst,
		}, nil
}

// Cleanup tries to clean the by-products of localize and logs if unsuccessful
func (ll *LocLoader) Cleanup() {
	cleanLdr(ll.ldr)
}

func cleanLdr(ldr ifc.Loader) {
	if err := ldr.Cleanup(); err != nil {
		log.Printf("%s", errors.WrapPrefixf(err, "unable to clean by-products of loading target").Error())
	}
}
