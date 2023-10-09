// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestDataValidation_NoName(t *testing.T) {
	fa := ConfigMapSecretFlagsAndArgs{}
	require.Error(t, fa.ValidateAdd([]string{}))
}

func TestDataValidation_MoreThanOneName(t *testing.T) {
	fa := ConfigMapSecretFlagsAndArgs{}

	require.Error(t, fa.ValidateAdd([]string{"name", "othername"}))
}

func TestDataConfigValidation_Flags(t *testing.T) {
	tests := []struct {
		name       string
		fa         ConfigMapSecretFlagsAndArgs
		shouldFail bool
	}{
		{
			name: "env-file-source and literal are both set",
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				EnvFileSource:  "three",
			},
			shouldFail: true,
		},
		{
			name: "env-file-source and from-file are both set",
			fa: ConfigMapSecretFlagsAndArgs{
				FileSources:   []string{"one", "two"},
				EnvFileSource: "three",
			},
			shouldFail: true,
		},
		{
			name:       "we don't have any option set",
			fa:         ConfigMapSecretFlagsAndArgs{},
			shouldFail: true,
		},
		{
			name: "we have from-file and literal ",
			fa: ConfigMapSecretFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				FileSources:    []string{"three", "four"},
			},
			shouldFail: false,
		},
		{
			name: "correct behavior",
			fa: ConfigMapSecretFlagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge",
			},
			shouldFail: false,
		},
		{
			name: "incorrect behavior",
			fa: ConfigMapSecretFlagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge-unknown",
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		err := test.fa.ValidateAdd([]string{"name"})
		if test.shouldFail {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestExpandFileSource(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	_, err := fSys.Create("dir/fa1")
	require.NoError(t, err)
	_, err = fSys.Create("dir/fa2")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"dir/fa*"},
	}
	err = fa.ExpandFileSource(fSys)
	require.NoError(t, err)
	expected := []string{
		"dir/fa1",
		"dir/fa2",
	}
	if !reflect.DeepEqual(fa.FileSources, expected) {
		t.Fatalf("FileSources is not correctly expanded: %v", fa.FileSources)
	}
}

func TestExpandFileSourceWithKey(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	_, err := fSys.Create("dir/faaaaaaaaaabbbbbbbbbccccccccccccccccc")
	require.NoError(t, err)
	_, err = fSys.Create("dir/foobar")
	require.NoError(t, err)
	_, err = fSys.Create("dir/simplebar")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*", "bar-key=dir/foobar", "dir/simplebar"},
	}
	err = fa.ExpandFileSource(fSys)
	require.NoError(t, err)
	expected := []string{
		"foo-key=dir/faaaaaaaaaabbbbbbbbbccccccccccccccccc",
		"bar-key=dir/foobar",
		"dir/simplebar",
	}
	if !reflect.DeepEqual(fa.FileSources, expected) {
		t.Fatalf("FileSources is not correctly expanded: %v", fa.FileSources)
	}
}

func TestExpandFileSourceWithKeyAndError(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	_, err := fSys.Create("dir/fa1")
	require.NoError(t, err)
	_, err = fSys.Create("dir/fa2")
	require.NoError(t, err)
	_, err = fSys.Create("dir/readme")
	require.NoError(t, err)
	fa := ConfigMapSecretFlagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*"},
	}
	err = fa.ExpandFileSource(fSys)
	if err == nil {
		t.Fatalf("FileSources should not be correctly expanded: %v", fa.FileSources)
	}
}
