package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/mod/modfile"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/utils"
)

const (
	dotDir = "."
)

// protoModule holds parts being collected to represent a module.
type protoModule struct {
	pathToGoMod string
	mf          *modfile.File
}

func (pm *protoModule) FullPath() string {
	return pm.mf.Module.Mod.Path
}

func (pm *protoModule) PathToGoMod() string {
	return pm.pathToGoMod
}

// Represents the trailing version label in a module name.
// See https://blog.golang.org/v2-go-modules
var trailingVersionPattern = regexp.MustCompile("/v\\d+$")

func (pm *protoModule) ShortName(
	repoImportPath string) misc.ModuleShortName {
	fp := pm.FullPath()
	if fp == repoImportPath {
		return misc.ModuleAtTop
	}
	p := fp[len(repoImportPath)+1:]
	stripped := trailingVersionPattern.ReplaceAllString(p, "")
	return misc.ModuleShortName(stripped)
}

func loadProtoModules(
	repoRoot string, exclusions []string) (result []*protoModule, err error) {
	var paths []string
	paths, err = getPathsToModules(repoRoot, exclusions)
	if err != nil {
		return
	}
	for _, p := range paths {
		var pm *protoModule
		pm, err = loadProtoModule(p)
		if err != nil {
			return
		}
		result = append(result, pm)
	}
	return
}

func loadProtoModule(path string) (*protoModule, error) {
	mPath := filepath.Join(path, goModFile)
	content, err := ioutil.ReadFile(mPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v\n", mPath, err)
	}
	f, err := modfile.Parse(mPath, content, nil)
	if err != nil {
		return nil, err
	}
	return &protoModule{pathToGoMod: path, mf: f}, nil
}

func getPathsToModules(
	repoRoot string, exclusions []string) (result []string, err error) {
	exclusionMap := utils.SliceToSet(exclusions)
	err = filepath.Walk(
		repoRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("trouble at pathToGoMod %q: %v\n", path, err)
			}
			if info.IsDir() {
				if _, ok := exclusionMap[info.Name()]; ok {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Name() == goModFile {
				result = append(result, path[:len(path)-len(goModFile)-1])
				return filepath.SkipDir
			}
			return nil
		})
	return
}
