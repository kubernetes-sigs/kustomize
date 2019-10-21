// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
)

func TestValidate(t *testing.T) {
	o := saveOptions{
		saveDirectory: "",
	}
	err := o.Validate()
	if !strings.Contains(err.Error(), "must specify one local directory") {
		t.Fatalf("Incorrect error %v", err)
	}

	o.saveDirectory = "/some/dir"
	err = o.Validate()
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
}

func TestComplete(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.Mkdir("/some/dir")
	fSys.WriteFile("/some/file", []byte(`some file`))

	type testcase struct {
		dir    string
		expect error
	}
	testcases := []testcase{
		{
			dir:    "/some/dir",
			expect: nil,
		},
		{
			dir:    "/some/dir/not/existing",
			expect: nil,
		},
		{
			dir:    "/some/file",
			expect: fmt.Errorf("%s is not a directory", "/some/file"),
		},
	}

	for _, tcase := range testcases {
		o := saveOptions{saveDirectory: tcase.dir}
		actual := o.Complete(fSys)
		if !reflect.DeepEqual(actual, tcase.expect) {
			t.Fatalf("Expected %v\n but bot %v\n", tcase.expect, actual)
		}
	}
}

func TestRunSave(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	o := saveOptions{saveDirectory: "/some/dir"}
	err := o.RunSave(fSys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !fSys.Exists("/some/dir/nameprefix.yaml") {
		t.Fatal("default configurations are not successfully save.")
	}
}
