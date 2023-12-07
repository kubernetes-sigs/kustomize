// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	v "github.com/spf13/viper"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/git"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/utils"
)

const (
	dotGitFileName = ".git"
	srcHint        = "/src/"
	goModFile      = "go.mod"
)

// DotGitData holds basic information about a local .git file
type DotGitData struct {
	// srcPath is the absolute path to the local Go src directory.
	// This used to be $GOPATH/src.
	// It's the directory containing git repository clones.
	srcPath string
	// The path below srcPath to a particular repository
	// directory, a directory containing a .git directory.
	// Typically {repoOrg}/{repoUserName}, e.g. sigs.k8s.io/cli-utils
	repoPath string
}

func (dg *DotGitData) SrcPath() string {
	return dg.srcPath
}

func (dg *DotGitData) RepoPath() string {
	return dg.repoPath
}

func (dg *DotGitData) AbsPath() string {
	return filepath.Join(dg.srcPath, dg.repoPath)
}

// NewDotGitDataFromPath wants the incoming path to hold dotGit
// E.g.
//
//	~/gopath/src/sigs.k8s.io/kustomize
//	~/gopath/src/github.com/monopole/gorepomod
func NewDotGitDataFromPath(path string, localFlag bool) (*DotGitData, error) {
	if !utils.DirExists(filepath.Join(path, dotGitFileName)) {
		return nil, fmt.Errorf(
			"%q doesn't have a %q file", path, dotGitFileName)
	}

	// If local flag is supplied, use local git naming instead of production (sigs.k8s.io)
	if localFlag {
		localPrefix, err := getLocalPrefix(path)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		v.Set("LocalGitPrefix", localPrefix)
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error extracting git repository %w", err)
		}

		pathSlice := strings.Split(wd, "/")

		return &DotGitData{
			srcPath:  strings.Join(pathSlice[:len(pathSlice)-1], "/"),
			repoPath: pathSlice[len(pathSlice)-1],
		}, nil
	} else {
		// This is an attempt to figure out where the user has cloned
		// their repos.  In the old days, it was an import path under
		// $GOPATH/src.  If we cannot guess it, we may need to ask for it,
		// or maybe proceed without knowing it.
		index := strings.Index(path, srcHint)
		if index < 0 {
			return nil, fmt.Errorf(
				"path %q doesn't contain %q", path, srcHint)
		}

		return &DotGitData{
			srcPath:  path[:index+len(srcHint)-1],
			repoPath: path[index+len(srcHint):],
		}, nil
	}
}

// It's a factory factory.
func (dg *DotGitData) NewRepoFactory(
	exclusions []string, localFlag bool) (*ManagerFactory, error) {
	modules, err := loadProtoModules(dg.AbsPath(), exclusions)

	if err != nil {
		return nil, err
	}

	if localFlag {
		err = dg.checkModulesLocal(modules)
		if err != nil {
			return nil, err
		}
	} else {
		err = dg.checkModules(modules)
		if err != nil {
			return nil, err
		}
	}

	runner := git.NewQuiet(dg.AbsPath(), true, localFlag)
	remoteName, err := runner.DetermineRemoteToUse()
	if err != nil {
		return nil, err
	}

	// Some tags might exist for modules that
	// have been renamed or deleted; ignore those.
	// There might be newer tags locally than remote,
	// so report both.
	localTags, err := runner.LoadLocalTags()
	if err != nil {
		return nil, err
	}
	remoteTags, err := runner.LoadRemoteTags(remoteName)
	if err != nil {
		return nil, err
	}

	return &ManagerFactory{
		dg:               dg,
		modules:          modules,
		remoteName:       remoteName,
		versionMapLocal:  localTags,
		versionMapRemote: remoteTags,
	}, nil
}

func (dg *DotGitData) checkModules(modules []*protoModule) error {
	for _, pm := range modules {

		file := filepath.Join(pm.PathToGoMod(), goModFile)

		// Do the paths make sense?
		if !strings.HasPrefix(pm.FullPath(), dg.RepoPath()) {
			return fmt.Errorf(
				"module %q doesn't start with the repository name %q",
				pm.FullPath(), dg.RepoPath())
		}

		shortName := pm.ShortName(dg.RepoPath())
		if shortName == misc.ModuleAtTop {
			if pm.PathToGoMod() != dg.AbsPath() {
				return fmt.Errorf("in %q, problem with top module", file)
			}
		} else {
			// Do the relative path and short name make sense?
			if !strings.HasSuffix(pm.PathToGoMod(), string(shortName)) {
				return fmt.Errorf(
					"in %q, the module name %q doesn't match the file's pathToGoMod %q",
					file, shortName, pm.PathToGoMod())
			}
		}
	}
	return nil
}

func (dg *DotGitData) checkModulesLocal(modules []*protoModule) error {
	for _, pm := range modules {
		file := filepath.Join(pm.PathToGoMod(), goModFile)
		_, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf(
				"cannot find go.mod file in %q", file)
		}
		pm.ShortNameWithLocalFlag(dg.RepoPath())
	}
	return nil
}

// Extract local prefix from origin url information retrieved from .git
func getLocalPrefix(dgAbsPath string) (string, error) {
	_, err := os.Stat(dgAbsPath + "/.git")
	if err != nil {
		return "", fmt.Errorf(".git directory does not exist in path %s", dgAbsPath)
	}

	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("failed extracting git information: %w", err)
	}

	localPrefix := utils.ParseGitRepositoryPath(string(out))
	if len(localPrefix) == 0 {
		_ = fmt.Errorf("parsed git repository path is empty: %w", err)
	}
	return localPrefix, nil
}
