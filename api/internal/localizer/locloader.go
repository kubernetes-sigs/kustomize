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

const dstPrefix = "localized"
const localizeDir = "localized-files"
const fatalPrefix = "should never happen"

const errAlpha = "alpha unsupported localize feature"
const errCleanDst = "unable to clean localize destination"
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
	IsRemote bool

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
	repo, isRemote := ldr.Repo()
	if isRemote {
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
	if isRemote {
		realScope = repo
	} else {
		realScope = scope.String()
	}
	// use target relative path to scope to find directory in destination corresponding to target
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
	errBadScopeForTarget := fmt.Sprintf("invalid localize scope '%s' for target '%s'", scopeArg, targetArg)

	if _, isRemote := targetLdr.Repo(); isRemote {
		if scopeArg != "" {
			return "", errors.WrapPrefixf(
				errors.Errorf("cannot specify scope for remote localize target"),
				errBadScopeForTarget)
		}
		return "", nil
	}
	// default scope
	if scopeArg == "" {
		return filesys.ConfirmedDir(targetLdr.Root()), nil
	}
	scope, err := filesys.ConfirmDir(fSys, scopeArg)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to establish localize scope")
	}
	if !filesys.ConfirmedDir(targetLdr.Root()).HasPrefix(scope) {
		return scope, errors.WrapPrefixf(errRootOutsideScope, errBadScopeForTarget)
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
	cleaned := strings.TrimRight(url, "/")
	i := strings.LastIndex(cleaned, "/")
	if i < 0 {
		return cleaned
	}
	return cleaned[i+1:]
}

// defaultNewDir calculates the default localize destination directory name from targetLdr at the localize target
// and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec) string {
	targetDir := filepath.Base(targetLdr.Root())
	if repo, isRemote := targetLdr.Repo(); isRemote {
		// kustomize doesn't download repo into repo-named folder
		// must find repo folder name from url
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		return strings.Join(
			[]string{dstPrefix, targetDir, strings.ReplaceAll(spec.Ref, "/", "-")},
			"-")
	}
	// special case for local target directory since destination directory cannot have "/" in name
	if targetDir == string(filepath.Separator) {
		return dstPrefix
	}
	return strings.Join([]string{dstPrefix, targetDir}, "-")
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
	// destination directory must sit in existing directory
	if err := fSys.Mkdir(newDirPath); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, newDirPath)
	if err != nil {
		if err := fSys.RemoveAll(newDir.String()); err != nil {
			log.Printf("%s", errors.WrapPrefixf(err, errCleanDst).Error())
		}
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

// Root returns the path to the current target directory
func (ll *LocLoader) Root() string {
	return ll.ldr.Root()
}

// Dst returns the path to the directory in destination that corresponds to Root
func (ll *LocLoader) Dst() string {
	return ll.dst
}

func toFSPath(urlPath string) string {
	return filepath.Join(strings.Split(urlPath, "/")...)
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
	if loader.HasRemoteFileScheme(path) {
		// parsable because prefixed by scheme
		u, err := url.Parse(path)
		if err != nil {
			log.Fatalf(errors.WrapPrefixf(
				err,
				"%s: cannot parse validated file url '%s'", fatalPrefix, path).Error())
		}
		return content, &LocPath{true, locFilePath(u)}, nil
	}
	if !filepath.IsAbs(path) {
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
		return content, &LocPath{false, pathSuffix(ll.ldr.Root(), cleanPath)}, nil
	}

	return nil, nil, errors.WrapPrefixf(errors.Errorf(errAlpha), "path '%s' is absolute", path)
}

func parseDomain(host string) string {
	if host == "gh:" {
		return "github.com"
	}
	target := host
	// remove scheme from target
	for _, p := range git.Schemes() {
		if strings.HasPrefix(target, p) {
			target = target[len(p):]
			break
		}
	}
	// remove userinfo
	if i := strings.Index(target, "@"); i > -1 {
		target = target[i+1:]
	}
	// remove ssh path delimiter or port delimiter
	if i := strings.Index(target, ":"); i > -1 {
		target = target[:i]
	}
	// remove http path delimiter
	return strings.TrimSuffix(target, "/")
}

// handleNewLocalDir returns the relative path between the current LocLoader root and the new root at ldr
// if the new root satisfies localize requirements. Otherwise, handleNewLocalDir returns an error.
func (ll *LocLoader) handleNewLocalDir(ldr ifc.Loader) (string, error) {
	if ll.local && !filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.scope) {
		return "", errors.WrapPrefixf(
			errRootOutsideScope,
			"invalid root location '%s' for scope '%s'", ldr.Root(), ll.args.scope)
	}
	if ll.local && filesys.ConfirmedDir(ldr.Root()).HasPrefix(ll.args.newDir) {
		return "", errors.WrapPrefixf(errDstRef, errReference, ldr.Root())
	}
	rel, err := filepath.Rel(ll.ldr.Root(), ldr.Root())
	if err != nil {
		log.Fatalf(errors.WrapPrefixf(err,
			"%s: cannot find relative path from root '%s' to its valid root reference '%s'",
			fatalPrefix, ll.ldr.Root(), ldr.Root()).Error())
	}
	return rel, nil
}

// New returns a LocLoader at path and whether path is remote if path is a root that can be localized.
// Otherwise, New returns an error.
func (ll *LocLoader) New(path string) (*LocLoader, *LocPath, error) {
	ldr, err := ll.ldr.New(path)
	if err != nil {
		return nil, nil, errors.WrapPrefixf(err, "invalid root reference")
	}
	var repo, inDst string
	var isRemote bool
	if repo, isRemote = ldr.Repo(); isRemote {
		spec, err := parseLocRootURL(path)
		if err != nil {
			return nil, nil, err
		}
		inRepo := pathSuffix(repo, ldr.Root())
		inDst = filepath.Join(localizeDir, parseDomain(spec.Host), toFSPath(spec.OrgRepo), toFSPath(spec.Ref), inRepo)
	} else {
		inDst, err = ll.handleNewLocalDir(ldr)
		if err != nil {
			return nil, nil, err
		}
	}
	return &LocLoader{
			fSys:  ll.fSys,
			args:  ll.args,
			ldr:   ldr,
			local: ll.local && !isRemote,
			dst:   filepath.Join(ll.dst, inDst),
		}, &LocPath{
			isRemote,
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
