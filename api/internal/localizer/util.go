// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const LocalizeDir = "localized-files"

// establishScope returns the scope given localize arguments and targetLdr at targetArg
func establishScope(scopeArg string, targetArg string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if repo, isRemote := targetLdr.Repo(); isRemote {
		if scopeArg != "" {
			return "", errors.Errorf("scope '%s' specified for remote localize target '%s'", scopeArg, targetArg)
		}
		return filesys.ConfirmedDir(repo), nil
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
		return scope, errors.Errorf("localize scope '%s' does not contain target '%s' at '%s'",
			scopeArg, targetArg, targetLdr.Root())
	}
	return scope, nil
}

// createNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func createNewDir(newDirArg string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if newDirArg == "" {
		newDirArg = defaultNewDir(targetLdr, spec)
	}
	if fSys.Exists(newDirArg) {
		return "", errors.Errorf("localize destination '%s' already exists", newDirArg)
	}
	// destination directory must sit in an existing directory
	if err := fSys.Mkdir(newDirArg); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, newDirArg)
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

// toFSPath returns urlPath as a file system path
func toFSPath(urlPath string) string {
	return filepath.Join(strings.Split(urlPath, "/")...)
}

// cleanFilePath returns the cleaned relPath, a validated relative file path inside root
func cleanFilePath(fSys filesys.FileSystem, root filesys.ConfirmedDir, file string) string {
	abs := root.Join(file)
	dir, f, err := fSys.CleanedAbs(abs)
	if err != nil {
		log.Fatalf("cannot clean validated file path '%s': %s", abs, err.Error())
	}
	locPath, err := filepath.Rel(root.String(), dir.Join(f))
	if err != nil {
		log.Fatalf("%s: %s", prefixRelErrWhenContains(root.String(), dir.Join(f)), err.Error())
	}
	return locPath
}

// locFilePath returns the relative localized path of validated file url u
func locFilePath(fileURL string) string {
	// file urls must have http or https scheme
	u, err := url.Parse(fileURL)
	if err != nil {
		log.Fatalf("cannot parse validated file url '%s': %s", fileURL, err.Error())
	}
	// raw github urls are the only type of file urls kustomize accepts;
	// path consists of org, repo, version, path in repo
	return filepath.Join(LocalizeDir, u.Hostname(), toFSPath(u.Path))
}

// locRootPath returns the relative localized path of the validated root url rootURL, where the local copy of its repo
// is at repoDir and the copy of its root is at rootDir
func locRootPath(rootURL string, repoDir filesys.ConfirmedDir, rootDir filesys.ConfirmedDir) string {
	repoSpec, err := git.NewRepoSpecFromURL(rootURL)
	if err != nil {
		log.Fatalf("cannot parse validated repo url '%s': %s", rootURL, err.Error())
	}
	// calculate from copy instead of url to straighten symlinks
	inRepo, err := filepath.Rel(repoDir.String(), rootDir.String())
	if err != nil {
		log.Fatalf("%s: %s", prefixRelErrWhenContains(repoDir.String(), rootDir.String()), err.Error())
	}
	return filepath.Join(LocalizeDir, parseDomain(repoSpec.Host), toFSPath(repoSpec.OrgRepo), toFSPath(repoSpec.Ref), inRepo)
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

// localizeGenerator localizes the paths in gen using lclzr
func localizeGenerator(gen *types.GeneratorArgs, lclzr *localizer) error {
	for i, env := range gen.EnvSources {
		newPath, err := lclzr.localizeFile(env)
		if err != nil {
			return err
		}
		gen.EnvSources[i] = newPath
	}
	for i, file := range gen.FileSources {
		k, f, err := kv.ParseFileSource(file)
		if err != nil {
			return errors.Wrap(err)
		}
		newFile, err := lclzr.localizeFile(f)
		if err != nil {
			return err
		}
		if file != f {
			newFile = k + "=" + newFile
		}
		gen.FileSources[i] = newFile
	}
	return nil
}
