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

package repourl

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewRepo(t *testing.T) {
	type testcase struct {
		repoUrl  string
		expected *Repo
	}

	testcases := []testcase{
		{
			repoUrl:  "git@github.com:kubernetes-sigs/kustomize.git",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "", ""},
		},
		{
			repoUrl:  "https://github.com/kubernetes-sigs/kustomize.git",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "", ""},
		},
		{
			repoUrl:  "https://github.com/kubernetes-sigs/kustomize",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "", ""},
		},
		{
			repoUrl:  "https://github.com/kubernetes-sigs/kustomize#v1.0.6",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "v1.0.6", ""},
		},
		{
			repoUrl:  "github.com/kubernetes-sigs/kustomize#017c4ae0aa19195db2a51ecc5aa82c56a1f1c99b",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "017c4ae0aa19195db2a51ecc5aa82c56a1f1c99b", ""},
		},
		{
			repoUrl:  "git@github.com:kubernetes-sigs/kustomize.git#test-branch",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "test-branch", ""},
		},
		{
			repoUrl:  "git@github.com:kubernetes-sigs/kustomize.git/examples/helloWorld",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "", "examples/helloWorld"},
		},
		{
			repoUrl:  "git@github.com:kubernetes-sigs/kustomize.git/examples/multibases#test-branch",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "test-branch", "examples/multibases"},
		},
		{
			repoUrl:  "https://github.com/kubernetes-sigs/kustomize/examples/multibases",
			expected: &Repo{"https://github.com/kubernetes-sigs/kustomize.git", "", "examples/multibases"},
		},
	}
	for _, tc := range testcases {
		actual, err := NewRepo(tc.repoUrl)
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("expected %v, but got %v", tc.expected, actual)
		}
	}
}

func TestNewRepoInvalid(t *testing.T) {
	invalidUrls := []string{
		"git@github.com/kubernetes-sigs//kustomize.git",
		"git@github.com:kubernetes-sigs//kustomize.git#bb#aa",
		"https://github.com//kubernetes-sigs/kustomize.git",
		"http://github.com/kubernetes-sigs/kustomize",
		"https://github.com/kubernetes-sigs/kustomize v1.0.6",
		"https@somehost.com/project/repo.git",
	}
	for _, url := range invalidUrls {
		_, err := NewRepo(url)
		if err == nil {
			t.Fatalf("Expected error happen %s", url)
		}
		if strings.Contains(err.Error(), "bababa") {
			t.Fatalf("unexpected error %s", err)
		}
	}
}
