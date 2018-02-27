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
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/kubectl/pkg/kinflate/util/fs"
)

func initializeRootLoader(fakefs fs.FileSystem) Loader {
	var schemes []SchemeLoader
	schemes = append(schemes, NewFileLoader(fakefs))
	rootLoader := Init(schemes)
	return rootLoader
}

func TestLoader_Root(t *testing.T) {

	// Initialize the fake file system and the root loader.
	fakefs := fs.MakeFakeFS()
	fakefs.WriteFile("/home/seans/project/file.yaml", []byte("Unused"))
	fakefs.WriteFile("/home/seans/project/subdir/file.yaml", []byte("Unused"))
	fakefs.WriteFile("/home/seans/project2/file.yaml", []byte("Unused"))
	rootLoader := initializeRootLoader(fakefs)

	_, err := rootLoader.New("")
	if err == nil {
		t.Fatalf("Expected error for empty root location not returned")
	}
	_, err = rootLoader.New("https://google.com/project")
	if err == nil {
		t.Fatalf("Expected error for unknown scheme not returned")
	}

	loader, err := rootLoader.New("/home/seans/project/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if "/home/seans/project" != loader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", loader.Root())
	}

	subLoader, err := loader.New("subdir/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if "/home/seans/project/subdir" != subLoader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", subLoader.Root())
	}

	anotherLoader, err := loader.New("/home/seans/project2/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if "/home/seans/project2" != anotherLoader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", anotherLoader.Root())
	}

	// Current directory should be expanded to a full absolute file path.
	currentDirLoader, err := loader.New(".")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if !filepath.IsAbs(currentDirLoader.Root()) {
		t.Fatalf("Incorrect Loader Root: %s\n", currentDirLoader.Root())
	}
}

func TestLoader_Load(t *testing.T) {

	// Initialize the fake file system and the root loader.
	fakefs := fs.MakeFakeFS()
	fakefs.WriteFile("/home/seans/project/file.yaml", []byte("This is a yaml file"))
	fakefs.WriteFile("/home/seans/project/subdir/file.yaml", []byte("Subdirectory file content"))
	fakefs.WriteFile("/home/seans/project2/file.yaml", []byte("This is another yaml file"))
	rootLoader := initializeRootLoader(fakefs)

	loader, err := rootLoader.New("/home/seans/project/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	fileBytes, err := loader.Load("file.yaml") // Load relative to root location
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual([]byte("This is a yaml file"), fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", "This is a yaml file", fileBytes)
	}

	fileBytes, err = loader.Load("subdir/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual([]byte("Subdirectory file content"), fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", "Subdirectory file content", fileBytes)
	}

	fileBytes, err = loader.Load("/home/seans/project2/file.yaml")
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual([]byte("This is another yaml file"), fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", "This is another yaml file", fileBytes)
	}

}
