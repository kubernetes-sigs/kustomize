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
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	// DstPrefix prefixes the target and ref, if target is remote, in the default localize destination directory name
	DstPrefix = "localized"

	// LocalizeDir is the name of the localize directories used to store remote content in the localize destination
	LocalizeDir = "localized-files"
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
		log.Fatalf("cannot parse validated file url %q: %s", fileURL, err.Error())
	}

	// Percent-encodings should be preserved in case sub-delims have special meaning.
	// Extraneous '..' parent directory dot-segments should be removed.
	path := filepath.Join(string(filepath.Separator), filepath.FromSlash(u.EscapedPath()))

	// The host should not include userinfo or port.
	// Raw github urls are the only type of file urls kustomize officially accepts.
	// In this case, the path already consists of org, repo, version, and path in repo, in order,
	// so we can use it as is.
	return filepath.Join(LocalizeDir, u.Hostname(), path)
}

// locRootPath returns the relative localized path of the validated root url rootURL, where the local copy of its repo
// is at repoDir and the copy of its root is at rootDir
// TODO(annasong): implement
func locRootPath(rootURL string, repoDir string, rootDir filesys.ConfirmedDir) string {
	_ = rootURL
	_, _ = repoDir, rootDir
	return ""
}

// localizeGenerator localizes the file paths on generator.
func localizeGenerator(lc *localizer, generator *types.GeneratorArgs) error {
	locEnvs := make([]string, len(generator.EnvSources))
	for i, env := range generator.EnvSources {
		newPath, err := lc.localizeFile(env)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator envs file")
		}
		locEnvs[i] = newPath
	}
	locFiles := make([]string, len(generator.FileSources))
	for i, file := range generator.FileSources {
		k, f, err := parseFileSource(file)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to parse generator files entry %q", file)
		}
		newFile, err := lc.localizeFile(f)
		if err != nil {
			return errors.WrapPrefixf(err, "unable to localize generator files path")
		}
		if f != file {
			newFile = k + "=" + newFile
		}
		locFiles[i] = newFile
	}
	generator.EnvSources = locEnvs
	generator.FileSources = locFiles
	return nil
}

// parseFileSource parses the source given.
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
		return "", "", errors.Errorf("key name for file path %v missing", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", errors.Errorf("file path for key name %v missing", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", errors.Errorf("key names or file paths cannot contain '='")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}
