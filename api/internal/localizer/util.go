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
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const LocalizeDir = "localized-files"

// establishScope returns the scope given localize arguments and targetLdr at rawTarget
func establishScope(rawScope string, rawTarget string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if repo, isRemote := targetLdr.Repo(); isRemote {
		if rawScope != "" {
			return "", errors.Errorf("scope %q specified for remote localize target %q", rawScope, rawTarget)
		}
		return filesys.ConfirmedDir(repo), nil
	}
	// default scope
	if rawScope == "" {
		return filesys.ConfirmedDir(targetLdr.Root()), nil
	}
	scope, err := filesys.ConfirmDir(fSys, rawScope)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to establish localize scope")
	}
	if !filesys.ConfirmedDir(targetLdr.Root()).HasPrefix(scope) {
		return scope, errors.Errorf("localize scope %q does not contain target %q at %q",
			rawScope, rawTarget, targetLdr.Root())
	}
	return scope, nil
}

// createNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func createNewDir(rawDst string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if rawDst == "" {
		rawDst = defaultNewDir(targetLdr, spec)
	}
	if fSys.Exists(rawDst) {
		return "", errors.Errorf("localize destination %q already exists", rawDst)
	}
	// destination directory must sit in an existing directory
	if err := fSys.Mkdir(rawDst); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, rawDst)
	if err != nil {
		if errCleanup := fSys.RemoveAll(newDir.String()); errCleanup != nil {
			log.Printf("%s", errors.WrapPrefixf(errCleanup, "unable to clean localize destination").Error())
		}
		return "", errors.WrapPrefixf(err, "unable to establish localize destination")
	}

	return newDir, nil
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
		return strings.Join([]string{dstPrefix, targetDir, strings.ReplaceAll(spec.Ref, "/", "-")}, "-")
	}
	// special case for local target directory since destination directory cannot have "/" in name
	if targetDir == string(filepath.Separator) {
		return dstPrefix
	}
	return strings.Join([]string{dstPrefix, targetDir}, "-")
}

// hasRef checks if path url has ref query string parameter
func hasRef(path string) bool {
	repoSpec, err := git.NewRepoSpecFromURL(path)
	if err != nil {
		log.Fatalf("%s: %s", "unable to parse validated root url", err.Error())
	}
	return repoSpec.Ref != ""
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

// cleanFilePath returns the cleaned relPath, a validated relative file path inside root
func cleanFilePath(fSys filesys.FileSystem, root filesys.ConfirmedDir, file string) string {
	abs := root.Join(file)
	dir, f, err := fSys.CleanedAbs(abs)
	if err != nil {
		log.Fatalf("cannot clean validated file path %q: %s", abs, err.Error())
	}
	locPath, err := filepath.Rel(root.String(), dir.Join(f))
	if err != nil {
		log.Fatalf("%s: %s", prefixRelErrWhenContains(root.String(), dir.Join(f)), err.Error())
	}
	return locPath
}

// locFilePath returns the relative localized path of validated file url fileURL
func locFilePath(fileURL string) string {
	// file urls must have http or https scheme
	u, err := url.Parse(fileURL)
	if err != nil {
		log.Fatalf("cannot parse validated file url %q: %s", fileURL, err.Error())
	}

	// preserve percent-encodings in case of sub-delims special meaning
	// remove extraneous '..' parent directory dot-segments
	path := filepath.Join(string(filepath.Separator), filepath.FromSlash(u.EscapedPath()))

	// raw github urls are the only type of file urls kustomize officially accepts, in which case
	// path consists of org, repo, version, path in repo

	// host should not include userinfo or port
	return filepath.Join(LocalizeDir, u.Hostname(), path)
}

// locRootPath returns the relative localized path of the validated root url rootURL, where the local copy of its repo
// is at repoDir and the copy of its root is at rootDir
func locRootPath(rootURL string, repoDir filesys.ConfirmedDir, rootDir filesys.ConfirmedDir) string {
	repoSpec, err := git.NewRepoSpecFromURL(rootURL)
	if err != nil {
		log.Fatalf("cannot parse validated repo url %q: %s", rootURL, err.Error())
	}
	// calculate from copy instead of url to straighten symlinks
	inRepo, err := filepath.Rel(repoDir.String(), rootDir.String())
	if err != nil {
		log.Fatalf("%s: %s", prefixRelErrWhenContains(repoDir.String(), rootDir.String()), err.Error())
	}
	// org, repo unlikely to contain dot-segments since repo is folder name when cloned
	// git does not allow ref value to contain dot-segments
	return filepath.Join(LocalizeDir,
		parseDomain(repoSpec.Host),
		filepath.FromSlash(repoSpec.OrgRepo),
		filepath.FromSlash(repoSpec.Ref),
		inRepo)
}

// parseDomain returns the domain from git.RepoSpec Host
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

// ParseFileSource parses the source given.
//
//	Acceptable formats include:
//	 1.  source-path: the basename will become the key name
//	 2.  source-name=source-path: the source-name will become the key name and
//	     source-path is the path to the key file.
//
// Key names cannot include '='.
func parseFileSource(source string) (keyName, filePath string, err error) {
	numSeparators := strings.Count(source, "=")
	switch {
	case numSeparators == 0:
		return filepath.Base(source), source, nil
	case numSeparators == 1 && strings.HasPrefix(source, "="):
		return "", "", fmt.Errorf("key name for file path %v missing", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", fmt.Errorf("file path for key name %v missing", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", errors.Errorf("key names or file paths cannot contain '='")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}

// localizeGenerator localizes the file paths on generator, which must not contain deprecated fields
func localizeGenerator(lc *Localizer, generator *types.GeneratorArgs) error {
	for i, env := range generator.EnvSources {
		newPath, err := lc.localizeFile(env)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator envs file")
		}
		generator.EnvSources[i] = newPath
	}
	for i, file := range generator.FileSources {
		k, f, err := parseFileSource(file)
		if err != nil {
			return errors.Wrap(err)
		}
		newFile, err := lc.localizeFile(f)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator files path")
		}
		if f != file {
			newFile = k + "=" + newFile
		}
		generator.FileSources[i] = newFile
	}
	return nil
}

// localizePatchesStrategicMerge localizes the file paths in patches
func localizePatchesStrategicMerge(lc *Localizer, patches []types.PatchStrategicMerge) error {
	for i := range patches {
		// try to parse as inline
		if _, err := lc.rFactory.RF().SliceFromBytes([]byte(patches[i])); err != nil {
			// must be file path
			newPath, err := lc.localizeFile(string(patches[i]))
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize patchesStrategicMerge path")
			}
			patches[i] = types.PatchStrategicMerge(newPath)
		}
	}
	return nil
}

// localizeReplacements localizes the file paths in replacements
func localizeReplacements(lc *Localizer, replacements []types.ReplacementField) error {
	for i := range replacements {
		if replacements[i].Path != "" {
			newPath, err := lc.localizeFile(replacements[i].Path)
			if err != nil {
				return errors.WrapPrefixf(err, "unable to localize replacements path")
			}
			replacements[i].Path = newPath
		}
	}
	return nil
}
