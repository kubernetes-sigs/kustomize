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

package add

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
)

func TestDataConfigValidation_NoName(t *testing.T) {
	config := cMapFlagsAndArgs{}

	if config.Validate([]string{}) == nil {
		t.Fatal("Validation should fail if no name is specified")
	}
}

func TestDataConfigValidation_MoreThanOneName(t *testing.T) {
	config := cMapFlagsAndArgs{}

	if config.Validate([]string{"name", "othername"}) == nil {
		t.Fatal("Validation should fail if more than one name is specified")
	}
}

func TestDataConfigValidation_Flags(t *testing.T) {
	tests := []struct {
		name       string
		config     cMapFlagsAndArgs
		shouldFail bool
	}{
		{
			name: "env-file-source and literal are both set",
			config: cMapFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				EnvFileSource:  "three",
			},
			shouldFail: true,
		},
		{
			name: "env-file-source and from-file are both set",
			config: cMapFlagsAndArgs{
				FileSources:   []string{"one", "two"},
				EnvFileSource: "three",
			},
			shouldFail: true,
		},
		{
			name:       "we don't have any option set",
			config:     cMapFlagsAndArgs{},
			shouldFail: true,
		},
		{
			name: "we have from-file and literal ",
			config: cMapFlagsAndArgs{
				LiteralSources: []string{"one", "two"},
				FileSources:    []string{"three", "four"},
			},
			shouldFail: false,
		},
	}

	for _, test := range tests {
		if test.config.Validate([]string{"name"}) == nil && test.shouldFail {
			t.Fatalf("Validation should fail if %s", test.name)
		} else if test.config.Validate([]string{"name"}) != nil && !test.shouldFail {
			t.Fatalf("Validation should succeed if %s", test.name)
		}
	}
}

func TestExpandFileSource(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.Create("dir/config1")
	fakeFS.Create("dir/config2")
	fakeFS.Create("dir/reademe")
	config := cMapFlagsAndArgs{
		FileSources: []string{"dir/config*"},
	}
	config.ExpandFileSource(fakeFS)
	expected := []string{
		"dir/config1",
		"dir/config2",
	}
	if !reflect.DeepEqual(config.FileSources, expected) {
		t.Fatalf("FileSources is not correctly expanded: %v", config.FileSources)
	}
}
