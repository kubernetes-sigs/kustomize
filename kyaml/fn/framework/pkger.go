// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/markbates/pkger"
)

type TemplatesFn func(*ResourceList) ([]*template.Template, error)

// TemplatesFromDir applies a directory of templates as generated resources.
func TemplatesFromDir(dirs ...pkger.Dir) TemplatesFn {
	return func(_ *ResourceList) ([]*template.Template, error) {
		var pt []*template.Template
		for i := range dirs {
			d := dirs[i]
			err := pkger.Walk(string(d), func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !strings.HasSuffix(info.Name(), ".template.yaml") {
					return nil
				}
				name := path.Join(string(d), info.Name())
				f, err := pkger.Open(name)
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				t, err := template.New(info.Name()).Parse(string(b))
				if err != nil {
					return err
				}

				pt = append(pt, t)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return pt, nil
	}
}

// PatchTemplatesFn returns a slice of PatchTemplate
type PatchTemplatesFn func(*ResourceList) ([]PatchTemplate, error)

// PT applies a directory of patches using the Selector
type PT struct {
	Selector func() *Selector
	Dir      pkger.Dir
}

// PatchTemplatesFromDir applies a directory of templates as patches.
func PatchTemplatesFromDir(templates ...PT) PatchTemplatesFn {
	return func(*ResourceList) ([]PatchTemplate, error) {
		var pt []PatchTemplate
		for i := range templates {
			v := templates[i]
			err := pkger.Walk(string(v.Dir), func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				name := path.Join(string(v.Dir), info.Name())

				if !strings.HasSuffix(info.Name(), ".template.yaml") {
					return nil
				}
				f, err := pkger.Open(name)
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				t, err := template.New(info.Name()).Parse(string(b))
				if err != nil {
					return err
				}

				pt = append(pt, PatchTemplate{Template: t, Selector: v.Selector()})
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return pt, nil
	}
}

// ContainerPatchTemplateFn returns a slice of ContainerPatchTemplate
type ContainerPatchTemplateFn func(*ResourceList) ([]ContainerPatchTemplate, error)

// CPT applies a directory of container patches using the Selector
type CPT struct {
	Selector func() *Selector
	Dir      pkger.Dir
	Names    []string
}

// ContainerPatchTemplatesFromDir applies a directory of templates as container patches.
func ContainerPatchTemplatesFromDir(templates ...CPT) ContainerPatchTemplateFn {
	return func(*ResourceList) ([]ContainerPatchTemplate, error) {
		var cpt []ContainerPatchTemplate
		for i := range templates {
			v := templates[i]
			err := pkger.Walk(string(v.Dir), func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !strings.HasSuffix(info.Name(), ".template.yaml") {
					return nil
				}

				name := path.Join(string(v.Dir), info.Name())
				f, err := pkger.Open(name)
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				t, err := template.New(info.Name()).Parse(string(b))
				if err != nil {
					return err
				}

				cpt = append(cpt, ContainerPatchTemplate{
					PatchTemplate:  PatchTemplate{Template: t, Selector: v.Selector()},
					ContainerNames: v.Names,
				})
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
		return cpt, nil
	}
}
