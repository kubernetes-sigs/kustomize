// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
)

func TestDataValidation_NoName(t *testing.T) {
	fa := flagsAndArgs{}

	if fa.Validate([]string{}) == nil {
		t.Fatal("Validation should fail if no name is specified")
	}
}

func TestDataValidation_MoreThanOneName(t *testing.T) {
	fa := flagsAndArgs{}

	if fa.Validate([]string{"name", "othername"}) == nil {
		t.Fatal("Validation should fail if more than one name is specified")
	}
}

func TestDataConfigValidation_Flags(t *testing.T) {
	tests := []struct {
		name       string
		fa         flagsAndArgs
		shouldFail bool
	}{
		{
			name: "env-file-source and literal are both set",
			fa: flagsAndArgs{
				LiteralSources: []string{"one", "two"},
				EnvFileSource:  "three",
			},
			shouldFail: true,
		},
		{
			name: "env-file-source and from-file are both set",
			fa: flagsAndArgs{
				FileSources:   []string{"one", "two"},
				EnvFileSource: "three",
			},
			shouldFail: true,
		},
		{
			name:       "we don't have any option set",
			fa:         flagsAndArgs{},
			shouldFail: true,
		},
		{
			name: "we have from-file and literal ",
			fa: flagsAndArgs{
				LiteralSources: []string{"one", "two"},
				FileSources:    []string{"three", "four"},
			},
			shouldFail: false,
		},
		{
			name: "correct behavior",
			fa: flagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge",
			},
			shouldFail: false,
		},
		{
			name: "incorrect behavior",
			fa: flagsAndArgs{
				EnvFileSource: "foo",
				Behavior:      "merge-unknown",
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		if test.fa.Validate([]string{"name"}) == nil && test.shouldFail {
			t.Fatalf("Validation should fail if %s", test.name)
		} else if test.fa.Validate([]string{"name"}) != nil && !test.shouldFail {
			t.Fatalf("Validation should succeed if %s", test.name)
		}
	}
}

func TestExpandFileSource(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.Create("dir/fa1")
	fSys.Create("dir/fa2")
	fSys.Create("dir/readme")
	fa := flagsAndArgs{
		FileSources: []string{"dir/fa*"},
	}
	fa.ExpandFileSource(fSys)
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
	fSys.Create("dir/faaaaaaaaaabbbbbbbbbccccccccccccccccc")
	fSys.Create("dir/foobar")
	fSys.Create("dir/simplebar")
	fSys.Create("dir/readme")
	fa := flagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*", "bar-key=dir/foobar", "dir/simplebar"},
	}
	fa.ExpandFileSource(fSys)
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
	fSys.Create("dir/fa1")
	fSys.Create("dir/fa2")
	fSys.Create("dir/readme")
	fa := flagsAndArgs{
		FileSources: []string{"foo-key=dir/fa*"},
	}
	err := fa.ExpandFileSource(fSys)
	if err == nil {
		t.Fatalf("FileSources should not be correctly expanded: %v", fa.FileSources)
	}
}
