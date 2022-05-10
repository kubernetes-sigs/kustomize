// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package parser_test

import (
	_ "embed"
	iofs "io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
)

//go:embed testdata/schema1.json
var schema1String string

//go:embed testdata/schema2.json
var schema2String string

func TestSchemaFiles(t *testing.T) {
	tests := []struct {
		name          string
		paths         []string
		fs            iofs.FS
		expectedCount int
		wantErr       string
	}{
		{
			name:          "parses schema from file",
			paths:         []string{"testdata/schema1.json"},
			expectedCount: 1,
		},
		{
			name:          "accepts multiple inputs",
			paths:         []string{"testdata/schema1.json", "testdata/schema2.json"},
			expectedCount: 2,
		},
		{
			name:          "parses templates from directory",
			paths:         []string{"testdata"},
			expectedCount: 2,
		},
		{
			name:          "can be configured with an alternative FS",
			fs:            os.DirFS("testdata"), // changes the root of the input paths
			paths:         []string{"schema1.json"},
			expectedCount: 1,
		},
		{
			name:    "rejects non-.template.yaml files",
			paths:   []string{"testdata/ignore.yaml"},
			wantErr: "file testdata/ignore.yaml does not have any of permitted extensions [.json]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.SchemaFiles(tt.paths...)
			if tt.fs != nil {
				p = p.FromFS(tt.fs)
			}
			schemas, err := p.Parse()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(schemas))
		})
	}
}

func TestSchemaStrings(t *testing.T) {
	tests := []struct {
		name          string
		data          []string
		expectedCount int
	}{
		{
			name:          "parses templates from strings",
			data:          []string{schema1String},
			expectedCount: 1,
		},
		{
			name:          "accepts multiple inputs",
			data:          []string{schema1String, schema2String},
			expectedCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.SchemaStrings(tt.data...)
			schemas, err := p.Parse()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(schemas))
		})
	}
}
