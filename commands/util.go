/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/ghodss/yaml"

	"k8s.io/kubectl/pkg/kustomize/constants"
	interror "k8s.io/kubectl/pkg/kustomize/internal/error"
	"k8s.io/kubectl/pkg/kustomize/types"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

type manifestFile struct {
	mPath string
	fsys  fs.FileSystem
}

func newManifestFile(mPath string, fsys fs.FileSystem) (*manifestFile, error) {
	mf := &manifestFile{mPath: mPath, fsys: fsys}
	err := mf.validate()
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *manifestFile) validate() error {
	f, err := mf.fsys.Stat(mf.mPath)
	if err != nil {
		errorMsg := fmt.Sprintf("Missing kustomize config file '%s'.\n", mf.mPath)
		merr := interror.ManifestError{ManifestFilepath: mf.mPath, ErrorMsg: errorMsg}
		return merr
	}
	if f.IsDir() {
		mf.mPath = path.Join(mf.mPath, constants.KustomizeFileName)
		_, err = mf.fsys.Stat(mf.mPath)
		if err != nil {
			errorMsg := fmt.Sprintf("Missing kustomize config file '%s'.\n", mf.mPath)
			merr := interror.ManifestError{ManifestFilepath: mf.mPath, ErrorMsg: errorMsg}
			return merr
		}
	} else {
		if !strings.HasSuffix(mf.mPath, constants.KustomizeFileName) {
			errorMsg := fmt.Sprintf("Kustomize config file path (%s) should have %s suffix\n", mf.mPath, constants.KustomizeSuffix)
			merr := interror.ManifestError{ManifestFilepath: mf.mPath, ErrorMsg: errorMsg}
			return merr
		}
	}
	return nil
}

func (mf *manifestFile) read() (*types.Manifest, error) {
	bytes, err := mf.fsys.ReadFile(mf.mPath)
	if err != nil {
		return nil, err
	}
	var manifest types.Manifest
	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, err
}

func (mf *manifestFile) write(manifest *types.Manifest) error {
	if manifest == nil {
		return errors.New("util: kustomize config file arg is nil.")
	}
	bytes, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return mf.fsys.WriteFile(mf.mPath, bytes)
}
