// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	"bytes"
	_ "embed"
	iofs "io/fs"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
)

//go:embed testdata/cm1.template.yaml
var cm1String string

//go:embed testdata/cm2.template.yaml
var cm2String string

var templateData = struct {
	Name string `yaml:"name"`
}{Name: "tester"}

var cm1Success = strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: appconfig
  labels:
    app: tester
data:
  app: tester
`)

var cm2Success = strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: env
  labels:
    app: tester
data:
  env: production
`)

func TestTemplateFiles(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		fs       iofs.FS
		expected []string
		wantErr  string
	}{
		{
			name:     "parses templates from file",
			paths:    []string{"testdata/cm1.template.yaml"},
			expected: []string{cm1Success},
		},
		{
			name:     "accepts multiple inputs",
			paths:    []string{"testdata/cm1.template.yaml", "testdata/cm2.template.yaml"},
			expected: []string{cm1Success, cm2Success},
		},
		{
			name:     "parses templates from directory",
			paths:    []string{"testdata"},
			expected: []string{cm1Success, cm2Success},
		},
		{
			name:     "can be configured with an alternative FS",
			fs:       os.DirFS("testdata"), // changes the root of the input paths
			paths:    []string{"cm1.template.yaml"},
			expected: []string{cm1Success},
		},
		{
			name:    "rejects non-.template.yaml files",
			paths:   []string{"testdata/ignore.yaml"},
			wantErr: "file testdata/ignore.yaml does not have any of permitted extensions [.template.yaml]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.TemplateFiles(tt.paths...)
			if tt.fs != nil {
				p = p.FromFS(tt.fs)
			}
			templates, err := p.Parse()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			result := []string{}
			for _, template := range templates {
				w := bytes.NewBuffer([]byte{})
				err := template.Execute(w, templateData)
				require.NoError(t, err)
				result = append(result, strings.TrimSpace(w.String()))
			}
			sort.Strings(tt.expected)
			sort.Strings(result)
			assert.Equal(t, len(result), len(tt.expected))
			for i := range tt.expected {
				assert.YAMLEq(t, tt.expected[i], result[i])
			}
		})
	}
}

func TestTemplateParser_WithExtensions(t *testing.T) {
	p := parser.TemplateFiles("testdata").WithExtensions(".json")
	templates, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, templates, 2)

	p = p.WithExtensions(".yaml")
	templates, err = p.Parse()
	require.NoError(t, err)
	require.Len(t, templates, 3)

	p = p.WithExtensions(".yaml", ".json")
	templates, err = p.Parse()
	require.NoError(t, err)
	require.Len(t, templates, 5)
}

func TestTemplateStrings(t *testing.T) {
	tests := []struct {
		name     string
		data     []string
		expected []string
	}{
		{
			name:     "parses templates from strings",
			data:     []string{cm1String},
			expected: []string{cm1Success},
		},
		{
			name:     "accepts multiple inputs",
			data:     []string{cm1String, cm2String},
			expected: []string{cm1Success, cm2Success},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.TemplateStrings(tt.data...)
			templates, err := p.Parse()
			require.NoError(t, err)

			result := []string{}
			for _, template := range templates {
				w := bytes.NewBuffer([]byte{})
				err := template.Execute(w, templateData)
				require.NoError(t, err)
				result = append(result, strings.TrimSpace(w.String()))
			}
			sort.Strings(tt.expected)
			sort.Strings(result)
			assert.Equal(t, len(result), len(tt.expected))
			for i := range tt.expected {
				assert.YAMLEq(t, tt.expected[i], result[i])
			}
		})
	}
}
