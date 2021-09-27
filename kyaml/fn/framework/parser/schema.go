// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	iofs "io/fs"
	"os"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

const (
	// SchemaExtension is the file extension this package requires schema files to have
	SchemaExtension = ".json"
)

// SchemaStrings returns a SchemaParser that will parse the schemas in the given strings.
//
// This is a helper for use with framework.TemplateProcessor#AdditionalSchemas. Example:
//
//	processor := framework.TemplateProcessor{
//		//...
//		AdditionalSchemas: parser.SchemaStrings(`
//        {
//          "definitions": {
//            "com.example.v1.Foo": {
//               ...
//            }
//          }
//        }
//		`),
//
func SchemaStrings(data ...string) framework.SchemaParser {
	return framework.SchemaParserFunc(func() ([]*spec.Definitions, error) {
		var defs []*spec.Definitions
		for _, content := range data {
			var schema spec.Schema
			if err := schema.UnmarshalJSON([]byte(content)); err != nil {
				return nil, err
			} else if schema.Definitions == nil {
				return nil, errors.Errorf("inline schema did not contain any definitions")
			}
			defs = append(defs, &schema.Definitions)
		}
		return defs, nil
	})
}

// SchemaFiles returns a SchemaParser that will parse the schemas in the given files.
// This is a helper for use with framework.TemplateProcessor#AdditionalSchemas.
//	processor := framework.TemplateProcessor{
//		//...
//		AdditionalSchemas: parser.SchemaFiles("path/to/crd-schemas", "path/to/special-schema.json),
//	}
func SchemaFiles(paths ...string) SchemaParser {
	return SchemaParser{parser{paths: paths, extension: SchemaExtension}}
}

// SchemaParser is a framework.SchemaParser that can parse files or directories containing openapi schemas.
type SchemaParser struct {
	parser
}

// Parse implements framework.SchemaParser
func (l SchemaParser) Parse() ([]*spec.Definitions, error) {
	if l.fs == nil {
		l.fs = os.DirFS(".")
	}

	var defs []*spec.Definitions
	err := l.parse(func(content []byte, name string) error {
		var schema spec.Schema
		if err := schema.UnmarshalJSON(content); err != nil {
			return err
		} else if schema.Definitions == nil {
			return errors.Errorf("schema %s did not contain any definitions", name)
		}
		defs = append(defs, &schema.Definitions)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return defs, nil
}

// FromFS allows you to replace the filesystem in which the parser will look up the given paths.
// For example, you can use an embed.FS.
func (l SchemaParser) FromFS(fs iofs.FS) SchemaParser {
	l.parser.fs = fs
	return l
}
