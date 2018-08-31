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

package transformer

import (
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/kubernetes-sigs/kustomize/pkg/internal/loadertest"
	"github.com/kubernetes-sigs/kustomize/pkg/patch"
)

func TestNewPatchJson6902FactoryNull(t *testing.T) {
	p := patch.PatchJson6902{
		Target: &patch.Target{
			Name: "some-name",
		},
	}
	f, err := NewPatchJson6902Factory(nil, p)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if f != nil {
		t.Fatal("a nil should be returned")
	}
}

func TestNewPatchJson6902FactoryNoTarget(t *testing.T) {
	p := patch.PatchJson6902{}
	_, err := NewPatchJson6902Factory(nil, p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "must specify the target field in patchesJson6902") {
		t.Fatalf("incorrect error returned: %v", err)
	}
}

func TestNewPatchJson6902FactoryConflict(t *testing.T) {
	jsonPatch := []byte(`
target:
  name: some-name
  kind: Deployment
jsonPatch:
  - op: replace
    path: /spec/template/spec/containers/0/name
    value: my-nginx
  - op: add
    path: /spec/template/spec/containers/0/command
    value: [arg1,arg2,arg3]
path: /some/dir/some/file
`)
	p := patch.PatchJson6902{}
	err := yaml.Unmarshal(jsonPatch, &p)
	if err != nil {
		t.Fatalf("expected error %v", err)
	}
	_, err = NewPatchJson6902Factory(nil, p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot specify path and jsonPath at the same time") {
		t.Fatalf("incorrect error returned %v", err)
	}
}

func TestNewPatchJson6902FactoryJSON(t *testing.T) {
	ldr := loadertest.NewFakeLoader("/testpath")
	operations := []byte(`[
        {"op": "replace", "path": "/spec/template/spec/containers/0/name", "value": "my-nginx"},
        {"op": "add", "path": "/spec/replica", "value": "3"},
        {"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["arg1", "arg2", "arg3"]}
]`)
	err := ldr.AddFile("/testpath/patch.json", operations)
	if err != nil {
		t.Fatalf("Failed to setup fake ldr.")
	}

	jsonPatch := []byte(`
target:
  kind: Deployment
  name: some-name
path: /testpath/patch.json
`)
	p := patch.PatchJson6902{}
	err = yaml.Unmarshal(jsonPatch, &p)
	if err != nil {
		t.Fatal("expected error")
	}

	f, err := NewPatchJson6902Factory(ldr, p)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if f == nil {
		t.Fatalf("the returned factory shouldn't be nil ")
	}

	_, err = f.MakePatchJson6902Transformer()
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
}

func TestNewPatchJson6902FactoryYAML(t *testing.T) {
	jsonPatch := []byte(`
target:
  name: some-name
  kind: Deployment
jsonPatch:
- op: replace
  path: /spec/template/spec/containers/0/name
  value: my-nginx
- op: add
  path: /spec/replica
  value: 3
- op: add
  path: /spec/template/spec/containers/0/command
  value: ["arg1", "arg2", "arg3"]
`)
	p := patch.PatchJson6902{}
	err := yaml.Unmarshal(jsonPatch, &p)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}

	f, err := NewPatchJson6902Factory(nil, p)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if f == nil {
		t.Fatalf("the returned factory shouldn't be nil ")
	}

	_, err = f.MakePatchJson6902Transformer()
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}

}
