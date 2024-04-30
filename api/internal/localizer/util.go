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
			targetDir = urlBase(spec.RepoPath)
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

// cleanedRelativePath returns a cleaned relative path of file to root on fSys
func cleanedRelativePath(fSys filesys.FileSystem, root filesys.ConfirmedDir, file string) string {
	abs := file
	if !filepath.IsAbs(file) {
		abs = root.Join(file)
	}

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

	// HTTP requests use the escaped path, so we use it here. Escaped paths also help us
	// preserve percent-encoding in the original path, in the absence of illegal characters,
	// in case they have special meaning to the host.
	// Extraneous '..' parent directory dot-segments should be removed.
	path := filepath.Join(string(filepath.Separator), filepath.FromSlash(u.EscapedPath()))

	// We intentionally exclude userinfo and port.
	// Raw github urls are the only type of file urls kustomize officially accepts.
	// In this case, the path already consists of org, repo, version, and path in repo, in order,
	// so we can use it as is.
	return filepath.Join(LocalizeDir, u.Hostname(), path)
}

// locRootPath returns the relative localized path of the validated root url rootURL, where the local copy of its repo
// is at repoDir and the copy of its root is at root on fSys.
func locRootPath(rootURL, repoDir string, root filesys.ConfirmedDir, fSys filesys.FileSystem) (string, error) {
	repoSpec, err := git.NewRepoSpecFromURL(rootURL)
	if err != nil {
		log.Panicf("cannot parse validated repo url %q: %s", rootURL, err)
	}
	host, err := parseHost(repoSpec)
	if err != nil {
		return "", errors.WrapPrefixf(err, "unable to parse host of remote root %q", rootURL)
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
	// the git-server-side directory name conventionally (but not universally) ends in .git, which
	// is conventionally stripped from the client-side directory name used for the clone.
	localRepoPath := strings.TrimSuffix(repoSpec.RepoPath, ".git")

	// We do not need to escape RepoPath, a path on the git server.
	// However, like git, we clean dot-segments from RepoPath.
	// Git does not allow ref value to contain dot-segments.
	return filepath.Join(LocalizeDir,
		host,
		filepath.Join(string(filepath.Separator), filepath.FromSlash(localRepoPath)),
		filepath.FromSlash(repoSpec.Ref),
		inRepo), nil
}

// parseHost returns the localize directory path corresponding to repoSpec.Host
func parseHost(repoSpec *git.RepoSpec) (string, error) {
	var target string
	switch scheme, _, _ := strings.Cut(repoSpec.Host, "://"); scheme {
	case "gh:":
		// 'gh' was meant to be a local github.com shorthand, in which case
		// the .gitconfig file could map it to any host. See origin here:
		// https://github.com/kubernetes-sigs/kustomize/blob/kustomize/v4.5.7/api/internal/git/repospec.go#L203
		// We give it a special host directory here under the assumption
		// that we are unlikely to have another host simply named 'gh'.
		return "gh", nil
	case "file":
		// We put file-scheme repos under a special directory to avoid
		// colluding local absolute paths with hosts.
		return FileSchemeDir, nil
	case "https", "http", "ssh":
		target = repoSpec.Host
	default:
		// We must have relative ssh url; in other words, the url has scp-like syntax.
		// We attach a scheme to avoid url.Parse errors.
		target = "ssh://" + repoSpec.Host
	}
	// url.Parse will not recognize ':' delimiter that both RepoSpec and git accept.
	target = strings.TrimSuffix(target, ":")
	u, err := url.Parse(target)
	if err != nil {
		return "", errors.Wrap(err)
	}
	// strip scheme, userinfo, port, and any trailing slashes.
	return u.Hostname(), nil
}
