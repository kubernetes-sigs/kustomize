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

package loader

import (
	"reflect"
	"testing"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

func TestLoader_Load(t *testing.T) {
	fakefs := fs.MakeFakeFS()
	location := "/home/seans/project/Kube-manifest.yaml"
	content := []byte("This is a kinflate manifest")
	fakefs.WriteFile(location, content)
	loader := RootLoader(location, fakefs)
	manifestBytes, err := loader.Load(location)
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual(content, manifestBytes) {
		t.Fatalf("expected %s, but got %s", content, manifestBytes)
	}
}
