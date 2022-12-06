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
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	// DstPrefix prefixes the target and ref, if target is remote, in the default localize destination directory name
	DstPrefix = "localized"

	// LocalizeDir is the name of the localize directories used to store remote content in the localize destination
	LocalizeDir = "localized-files"

	// FileSchemeDir is the name of the directory immediately inside LocalizeDir used to store file-schemed repos
	FileSchemeDir = "file-schemed"
)

// establishScope returns the effective scope given localize arguments and targetLdr at rawTarget. For remote rawTarget,
// the effective scope is the downloaded repo.
func establishScope(rawScope string, rawTarget string, targetLdr ifc.Loader, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if repo := targetLdr.Repo(); repo != "" {
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
		return scope, errors.Errorf("localize scope %q does not contain target %q at %q", rawScope, rawTarget,
			targetLdr.Root())
	}
	return scope, nil
}

// createNewDir returns the localize destination directory or error. Note that spec is nil if targetLdr is at local
// target.
func createNewDir(rawNewDir string, targetLdr ifc.Loader, spec *git.RepoSpec, fSys filesys.FileSystem) (filesys.ConfirmedDir, error) {
	if rawNewDir == "" {
		rawNewDir = defaultNewDir(targetLdr, spec)
	}
	if fSys.Exists(rawNewDir) {
		return "", errors.Errorf("localize destination %q already exists", rawNewDir)
	}
	// destination directory must sit in an existing directory
	if err := fSys.Mkdir(rawNewDir); err != nil {
		return "", errors.WrapPrefixf(err, "unable to create localize destination directory")
	}
	newDir, err := filesys.ConfirmDir(fSys, rawNewDir)
	if err != nil {
		if errCleanup := fSys.RemoveAll(newDir.String()); errCleanup != nil {
			log.Printf("%s", errors.WrapPrefixf(errCleanup, "unable to clean localize destination"))
		}
		return "", errors.WrapPrefixf(err, "unable to establish localize destination")
	}

	return newDir, nil
}

// defaultNewDir calculates the default localize destination directory name from targetLdr at the localize target
// and spec of target, which is nil if target is local
func defaultNewDir(targetLdr ifc.Loader, spec *git.RepoSpec) string {
	targetDir := filepath.Base(targetLdr.Root())
	if repo := targetLdr.Repo(); repo != "" {
		// kustomize doesn't download repo into repo-named folder
		// must find repo folder name from url
		if repo == targetLdr.Root() {
			targetDir = urlBase(spec.OrgRepo)
		}
		return strings.Join([]string{DstPrefix, targetDir, strings.ReplaceAll(spec.Ref, "/", "-")}, "-")
	}
	// special case for local target directory since destination directory cannot have "/" in name
	if targetDir == string(filepath.Separator) {
		return DstPrefix
	}
	return strings.Join([]string{DstPrefix, targetDir}, "-")
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

// hasRef checks if repoURL has ref query string parameter
func hasRef(repoURL string) bool {
	repoSpec, err := git.NewRepoSpecFromURL(repoURL)
	if err != nil {
		log.Fatalf("unable to parse validated root url: %s", err)
	}
	return repoSpec.Ref != ""
}

// cleanFilePath returns file cleaned, where file is a relative path to root on fSys
func cleanFilePath(fSys filesys.FileSystem, root filesys.ConfirmedDir, file string) string {
	abs := root.Join(file)
	dir, f, err := fSys.CleanedAbs(abs)
	if err != nil {
		log.Fatalf("cannot clean validated file path %q: %s", abs, err)
	}
	locPath, err := filepath.Rel(root.String(), dir.Join(f))
	if err != nil {
		log.Fatalf("cannot find path from parent %q to file %q: %s", root, dir.Join(f), err)
	}
	return locPath
}

// locFilePath converts a URL to its localized form, e.g.
// https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/service.yaml ->
// localized-files/raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/service.yaml.
//
// fileURL must be a validated file URL.
func locFilePath(fileURL string) string {
	// File urls must have http or https scheme, so it is safe to use url.Parse.
	u, err := url.Parse(fileURL)
	if err != nil {
		log.Panicf("cannot parse validated file url %q: %s", fileURL, err)
	}

	// We include both userinfo and port, lest they determine the file read.
	authority := strings.TrimPrefix((&url.URL{
		User: u.User,
		Host: u.Host,
	}).String(), "//")

	// Extraneous '..' parent directory dot-segments should be removed.
	path := filepath.Join(string(filepath.Separator), filepath.FromSlash(u.EscapedPath()))

	// Raw github urls are the only type of file urls kustomize officially accepts.
	// In this case, the path already consists of org, repo, version, and path in repo, in order,
	// so we can use it as is.
	return filepath.Join(LocalizeDir, authority, path)
}

// locRootPath returns the relative localized path of the validated root url rootURL, where the local copy of its repo
// is at repoDir and the copy of its root is at root on fSys.
func locRootPath(rootURL string, repoDir string, root filesys.ConfirmedDir, fSys filesys.FileSystem) string {
	repoSpec, err := git.NewRepoSpecFromURL(rootURL)
	if err != nil {
		log.Panicf("cannot parse validated repo url %q: %s", rootURL, err)
	}
	repo, err := filesys.ConfirmDir(fSys, repoDir)
	if err != nil {
		log.Panicf("unable to establish validated repo download location %q: %s", repoDir, err)
	}
	// calculate from copy instead of url to straighten symlinks
	inRepo, err := filepath.Rel(repo.String(), root.String())
	if err != nil {
		log.Panicf("cannot find path from %q to child directory %q: %s", repo, root, err)
	}
	// Like git, we clean dot-segments from OrgRepo.
	// Git does not allow ref value to contain dot-segments.
	return filepath.Join(LocalizeDir,
		parseSchemeAuthority(repoSpec),
		filepath.Join(string(filepath.Separator), filepath.FromSlash(repoSpec.OrgRepo)),
		filepath.FromSlash(repoSpec.Ref),
		inRepo)
}

// parseSchemeAuthority returns the localize directory path corresponding to repoSpec.Host
func parseSchemeAuthority(repoSpec *git.RepoSpec) string {
	if repoSpec.Host == "file://" {
		return FileSchemeDir
	}
	target := repoSpec.Host
	for _, scheme := range []string{"https", "http", "ssh"} {
		schemePrefix := scheme + "://"
		if strings.HasPrefix(target, schemePrefix) {
			target = target[len(schemePrefix):]
			break
		}
	}
	// We remove delimiters in this order, as ':' has different meaning if prefixed by '/'.
	// We can remove ':/' suffix because ':' delimits empty port in this case.
	// Note that gh: is handled here.
	target = strings.TrimSuffix(target, "/")
	return strings.TrimSuffix(target, ":")
}
