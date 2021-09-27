// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

const (
	// TemplateExtension is the file extension this package requires template files to have
	TemplateExtension = ".template.yaml"
)

// TemplateStrings returns a TemplateParser that will parse the templates from the given strings.
//
// This is a helper for use with framework.TemplateProcessor's template subfields. Example:
//
//	 processor := framework.TemplateProcessor{
//		ResourceTemplates: []framework.ResourceTemplate{{
//			Templates: parser.TemplateStrings(`
//				apiVersion: apps/v1
//				kind: Deployment
//				metadata:
//				 name: foo
//				 namespace: default
//				 annotations:
//				   {{ .Key }}: {{ .Value }}
//				`)
//		}},
//	 }
func TemplateStrings(data ...string) framework.TemplateParser {
	return framework.TemplateParserFunc(func() ([]*template.Template, error) {
		var templates []*template.Template
		for i := range data {
			t, err := template.New(fmt.Sprintf("inline%d", i)).Parse(data[i])
			if err != nil {
				return nil, err
			}
			templates = append(templates, t)
		}
		return templates, nil
	})
}

// TemplateFiles returns a TemplateParser that will parse the templates from the given files or directories.
// Only immediate children of any given directories will be parsed.
// All files must end in .template.yaml.
//
// This is a helper for use with framework.TemplateProcessor's template subfields. Example:
//
// 	 processor := framework.TemplateProcessor{
//		ResourceTemplates: []framework.ResourceTemplate{{
//			Templates: parser.TemplateFiles("path/to/templates", "path/to/special.template.yaml")
//		}},
//   }
func TemplateFiles(paths ...string) TemplateParser {
	return TemplateParser{parser{paths: paths, extension: TemplateExtension}}
}

// TemplateParser is a framework.TemplateParser that can parse files or directories containing Go templated YAML.
type TemplateParser struct {
	parser
}

// Parse implements framework.TemplateParser
func (l TemplateParser) Parse() ([]*template.Template, error) {
	if l.fs == nil {
		l.fs = os.DirFS(".")
	}

	var templates []*template.Template
	err := l.parse(func(content []byte, file string) error {
		t, err := template.New(filepath.Base(file)).Parse(string(content))
		if err == nil {
			templates = append(templates, t)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// FromFS allows you to replace the filesystem in which the parser will look up the given paths.
// For example, you can use an embed.FS.
func (l TemplateParser) FromFS(fs iofs.FS) TemplateParser {
	l.parser.fs = fs
	return l
}
