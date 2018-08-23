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
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	interror "github.com/kubernetes-sigs/kustomize/pkg/internal/error"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

var (
	// These field names are the exact kustomization fields
	kustomizationFields = []string{
		"APIVersion",
		"Kind",
		"Resources",
		"Bases",
		"NamePrefix",
		"Namespace",
		"Crds",
		"CommonLabels",
		"CommonAnnotations",
		"Patches",
		"ConfigMapGenerator",
		"SecretGenerator",
		"Vars",
		"ImageTags",
	}
)

// commentedField records the comment associated with a kustomization field
// field has to be a recognized kustomization field
// comment can be empty
type commentedField struct {
	field   string
	comment []byte
}

func (cf *commentedField) appendComment(comment []byte) {
	cf.comment = append(cf.comment, comment...)
}

func squash(x [][]byte) []byte {
	return bytes.Join(x, []byte(``))
}

type kustomizationFile struct {
	path           string
	fsys           fs.FileSystem
	originalFields []*commentedField
}

func newKustomizationFile(mPath string, fsys fs.FileSystem) (*kustomizationFile, error) { // nolint
	mf := &kustomizationFile{path: mPath, fsys: fsys}
	err := mf.validate()
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *kustomizationFile) validate() error {
	if !mf.fsys.Exists(mf.path) {
		errorMsg := fmt.Sprintf("Missing kustomization file '%s'.\n", mf.path)
		merr := interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
		return merr
	}
	if mf.fsys.IsDir(mf.path) {
		mf.path = path.Join(mf.path, constants.KustomizationFileName)
		if !mf.fsys.Exists(mf.path) {
			errorMsg := fmt.Sprintf("Missing kustomization file '%s'.\n", mf.path)
			merr := interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
			return merr
		}
	} else {
		if !strings.HasSuffix(mf.path, constants.KustomizationFileName) {
			errorMsg := fmt.Sprintf("Kustomization file path (%s) should have %s suffix\n",
				mf.path, constants.KustomizationFileSuffix)
			return interror.KustomizationError{KustomizationPath: mf.path, ErrorMsg: errorMsg}
		}
	}
	return nil
}

func (mf *kustomizationFile) read() (*types.Kustomization, error) {
	data, err := mf.fsys.ReadFile(mf.path)
	if err != nil {
		return nil, err
	}
	var kustomization types.Kustomization
	err = yaml.Unmarshal(data, &kustomization)
	if err != nil {
		return nil, err
	}
	err = mf.parseCommentedFields(data)
	if err != nil {
		return nil, err
	}
	return &kustomization, err
}

func (mf *kustomizationFile) write(kustomization *types.Kustomization) error {
	if kustomization == nil {
		return errors.New("util: kustomization file arg is nil")
	}
	data, err := mf.marshal(kustomization)
	if err != nil {
		return err
	}
	return mf.fsys.WriteFile(mf.path, data)
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func (mf *kustomizationFile) parseCommentedFields(content []byte) error {
	buffer := bytes.NewBuffer(content)
	var comments [][]byte

	line, err := buffer.ReadBytes('\n')
	for err == nil {
		if isCommentOrBlankLine(line) {
			comments = append(comments, line)
		} else {
			matched, field := findMatchedField(line)
			if matched {
				mf.originalFields = append(mf.originalFields, &commentedField{field: field, comment: squash(comments)})
				comments = [][]byte{}
			} else if len(comments) > 0 {
				mf.originalFields[len(mf.originalFields)-1].appendComment(squash(comments))
				comments = [][]byte{}
			}
		}
		line, err = buffer.ReadBytes('\n')
	}

	if err != io.EOF {
		return err
	}
	return nil
}

func (mf *kustomizationFile) marshal(kustomization *types.Kustomization) ([]byte, error) {
	var output []byte
	for _, comment := range mf.originalFields {
		output = append(output, comment.comment...)
		content, err := marshalField(comment.field, kustomization)
		if err != nil {
			return content, err
		}
		output = append(output, content...)
	}
	for _, field := range kustomizationFields {
		if mf.hasField(field) {
			continue
		}
		content, err := marshalField(field, kustomization)
		if err != nil {
			return content, nil
		}
		output = append(output, content...)

	}
	return output, nil
}

func (mf *kustomizationFile) hasField(name string) bool {
	for _, n := range mf.originalFields {
		if n.field == name {
			return true
		}
	}
	return false
}

/*
 isCommentOrBlankLine determines if a line is a comment or blank line
 Return true for following lines
 # This line is a comment
       # This line is also a comment with several leading white spaces

 (The line above is a blank line)
*/
func isCommentOrBlankLine(line []byte) bool {
	s := bytes.TrimRight(bytes.TrimLeft(line, " "), "\n")
	return len(s) == 0 || bytes.HasPrefix(s, []byte(`#`))
}

func findMatchedField(line []byte) (bool, string) {
	for _, field := range kustomizationFields {
		// (?i) is for case insensitive regexp matching
		r := regexp.MustCompile("^(" + "(?i)" + field + "):")
		if r.Match(line) {
			return true, field
		}
	}
	return false, ""
}

// marshalField marshal a given field of a kustomization object into yaml format.
// If the field wasn't in the original kustomization.yaml file or wasn't added,
// an empty []byte is returned.
func marshalField(field string, kustomization *types.Kustomization) ([]byte, error) {
	r := reflect.ValueOf(*kustomization)
	v := r.FieldByName(strings.Title(field))

	if !v.IsValid() || v.Len() == 0 {
		return []byte{}, nil
	}

	k := &types.Kustomization{}
	kr := reflect.ValueOf(k)
	kv := kr.Elem().FieldByName(strings.Title(field))
	kv.Set(v)

	return yaml.Marshal(k)
}
