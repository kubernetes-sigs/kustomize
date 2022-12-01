// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/generators"
)

func TestParseFileSource(t *testing.T) {
	tests := map[string]*struct {
		Input    string
		Error    string
		Key      string
		Filename string
	}{
		"filename only": {
			Input:    "./path/myfile",
			Key:      "myfile",
			Filename: "./path/myfile",
		},
		"key and filename": {
			Input:    "newName.ini=oldName",
			Key:      "newName.ini",
			Filename: "oldName",
		},
		"multiple =": {
			Input: "newName.ini==oldName",
			Error: `source "newName.ini==oldName" key name or file path contains '='`,
		},
		"missing key": {
			Input: "=myfile",
			Error: `missing key name for file path "myfile" in source "=myfile"`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			key, file, err := ParseFileSource(test.Input)
			if test.Error != "" {
				require.EqualError(t, err, test.Error)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.Key, key)
				require.Equal(t, test.Filename, file)
			}
		})
	}
}
