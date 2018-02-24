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

var rootDir = "/home/seans/project"
var rootFile = "file.yaml"
var rootFilePath = filepath.Join(rootDir, rootFile)
var rootFileContent = []byte("This is a yaml file")

var subDirectory = "subdir"
var subDirectoryPath = filepath.Join(subDirectory, rootFile)
var subDirectoryContent = []byte("Subdirectory file content")

var anotherRootDir = "/home/seans/project2"
var anotherFilePath = filepath.Join(anotherRootDir, rootFile)
var anotherFileContent = []byte("This is another yaml file")

func TestLoader_Root(t *testing.T) {

	rootLoader := initializeRootLoader()
	_, err := rootLoader.New("")
	if err == nil {
		t.Fatalf("Expected error for empty root location not returned")
	}
	_, err = rootLoader.New("https://google.com/project")
	if err == nil {
		t.Fatalf("Expected error for unknown scheme not returned")
	}

	loader, err := rootLoader.New(rootFilePath)
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if rootDir != loader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", loader.Root())
	}

	subLoader, err := loader.New(subDirectoryPath)
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if filepath.Join(rootDir, subDirectory) != subLoader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", subLoader.Root())
	}

	anotherLoader, err := loader.New(anotherFilePath)
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if anotherRootDir != anotherLoader.Root() {
		t.Fatalf("Incorrect Loader Root: %s\n", anotherLoader.Root())
	}

	currentDirLoader, err := loader.New(".")
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	if !filepath.IsAbs(currentDirLoader.Root()) {
		t.Fatalf("Incorrect Loader Root: %s\n", currentDirLoader.Root())
	}
}

func TestLoader_Load(t *testing.T) {
	rootLoader := initializeRootLoader()
	loader, err := rootLoader.New(rootFilePath)
	if err != nil {
		t.Fatalf("Unexpected in New(): %v\n", err)
	}
	fileBytes, err := loader.Load(rootFilePath)
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual(rootFileContent, fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", rootFileContent, fileBytes)
	}
	fileBytes, err = loader.Load(subDirectoryPath)
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual(subDirectoryContent, fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", subDirectoryContent, fileBytes)
	}
	fileBytes, err = loader.Load(anotherFilePath)
	if err != nil {
		t.Fatalf("Unexpected error in Load(): %v", err)
	}
	if !reflect.DeepEqual(anotherFileContent, fileBytes) {
		t.Fatalf("Load failed. Expected %s, but got %s", anotherFileContent, fileBytes)
	}

}

func initializeRootLoader() Loader {
	fs := initializeFakeFilesystem()
	var schemes []SchemeLoader
	schemes = append(schemes, NewFileLoader(fs))
	rootLoader := Init(schemes)
	return rootLoader
}

func initializeFakeFilesystem() fs.FileSystem {
	fakefs := fs.MakeFakeFS()
	fakefs.WriteFile(rootFilePath, rootFileContent)
	fakefs.WriteFile(filepath.Join(rootDir, subDirectoryPath), subDirectoryContent)
	fakefs.WriteFile(anotherFilePath, anotherFileContent)
	return fakefs
}
