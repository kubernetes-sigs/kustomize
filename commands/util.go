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

type kustomizationFile struct {
	path string
	fsys fs.FileSystem
}

func newKustomizationFile(mPath string, fsys fs.FileSystem) (*kustomizationFile, error) {
	mf := &kustomizationFile{path: mPath, fsys: fsys}
	err := mf.validate()
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *kustomizationFile) validate() error {
	f, err := mf.fsys.Stat(mf.path)
	if err != nil {
		errorMsg := fmt.Sprintf("Missing kustomization file '%s'.\n", mf.path)
		merr := interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
		return merr
	}
	if f.IsDir() {
		mf.path = path.Join(mf.path, constants.KustomizationFileName)
		_, err = mf.fsys.Stat(mf.path)
		if err != nil {
			errorMsg := fmt.Sprintf("Missing kustomization file '%s'.\n", mf.path)
			merr := interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
			return merr
		}
	} else {
		if !strings.HasSuffix(mf.path, constants.KustomizationFileName) {
			errorMsg := fmt.Sprintf("Kustomization file path (%s) should have %s suffix\n", mf.path, constants.KustomizationSuffix)
			merr := interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
			return merr
		}
	}
	return nil
}

func (mf *kustomizationFile) read() (*types.Kustomization, error) {
	bytes, err := mf.fsys.ReadFile(mf.path)
	if err != nil {
		return nil, err
	}
	var kustomization types.Kustomization
	err = yaml.Unmarshal(bytes, &kustomization)
	if err != nil {
		return nil, err
	}
	return &kustomization, err
}

func (mf *kustomizationFile) write(kustomization *types.Kustomization) error {
	if kustomization == nil {
		return errors.New("util: kustomization file arg is nil.")
	}
	bytes, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}

	return mf.fsys.WriteFile(mf.path, bytes)
}
