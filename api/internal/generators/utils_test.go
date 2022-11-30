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
		Legal    bool
		Key      string
		Filename string
	}{
		"filename only": {
			"./path/myfile",
			true,
			"myfile",
			"./path/myfile",
		},
		"key and filename": {
			"newName.ini=oldName",
			true,
			"newName.ini",
			"oldName",
		},
		"multiple =": {
			Input: "newName.ini==oldName",
		},
		"missing key": {
			Input: "=myfile",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			key, file, err := ParseFileSource(test.Input)
			if !test.Legal {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.Key, key)
				require.Equal(t, test.Filename, file)
			}
		})
	}
}
