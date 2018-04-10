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

package fs

import (
	"bytes"
	"testing"
)

func TestStatNotExist(t *testing.T) {
	x := MakeFakeFS()
	info, err := x.Stat("foo")
	if info != nil {
		t.Fatalf("expected nil info")
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestStat(t *testing.T) {
	x := MakeFakeFS()
	expectedName := "my-dir"
	err := x.Mkdir(expectedName, 0666)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := x.Stat(expectedName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := info.Name()
	if name != expectedName {
		t.Fatalf("expected %v but got %v", expectedName, name)
	}
	if !info.IsDir() {
		t.Fatalf("expected IsDir() return true")
	}
}

func TestCreate(t *testing.T) {
	x := MakeFakeFS()
	f, err := x.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	info, err := x.Stat("foo")
	if info == nil {
		t.Fatalf("expected non-nil info")
	}
	if err != nil {
		t.Fatalf("expected no error")
	}
}

func TestReadFile(t *testing.T) {
	x := MakeFakeFS()
	f, err := x.Create("foo")
	if f == nil {
		t.Fatalf("expected file")
	}
	if err != nil {
		t.Fatalf("unexpected error")
	}
	content, err := x.ReadFile("foo")
	if len(content) != 0 {
		t.Fatalf("expected no content")
	}
	if err != nil {
		t.Fatalf("expected no error")
	}
}

func TestWriteFile(t *testing.T) {
	x := MakeFakeFS()
	c := []byte("heybuddy")
	err := x.WriteFile("foo", c)
	if err != nil {
		t.Fatalf("expected no error")
	}
	content, err := x.ReadFile("foo")
	if err != nil {
		t.Fatalf("expected read to work: %v", err)
	}
	if bytes.Compare(c, content) != 0 {
		t.Fatalf("incorrect content: %v", content)
	}
}
