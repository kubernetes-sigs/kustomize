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

package types

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"testing"
)

func TestGVK(t *testing.T) {
	type testcase struct {
		data     string
		expected gvk.Gvk
	}

	testcases := []testcase{
		{
			data: `
apiVersion: v1
kind: Secret
name: my-secret
`,
			expected: gvk.Gvk{Group: "", Version: "v1", Kind: "Secret"},
		},
		{
			data: `
apiVersion: myapps/v1
kind: MyKind
name: my-kind
`,
			expected: gvk.Gvk{Group: "myapps", Version: "v1", Kind: "MyKind"},
		},
		{
			data: `
version: v2
kind: MyKind
name: my-kind
`,
			expected: gvk.Gvk{Version: "v2", Kind: "MyKind"},
		},
	}

	for _, tc := range testcases {
		var targ Target
		err := yaml.Unmarshal([]byte(tc.data), &targ)
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}
		if !reflect.DeepEqual(targ.GVK(), tc.expected) {
			t.Fatalf("Expected %v, but got %v", tc.expected, targ.GVK())
		}
	}
}

func TestDefaulting(t *testing.T) {
	v := &Var{
		Name: "SOME_VARIABLE_NAME",
		ObjRef: Target{
			Gvk: gvk.Gvk{
				Version: "v1",
				Kind:    "Secret",
			},
			Name: "my-secret",
		},
	}
	v.Defaulting()
	if v.FieldRef.FieldPath != "metadata.name" {
		t.Fatalf("var defaulting doesn't behave correctly.\n expected metadata.name, but got %v", v.FieldRef.FieldPath)
	}
}
