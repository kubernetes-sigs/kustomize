// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

func TestMerge3_Merge(t *testing.T) {
	_, datadir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	datadir = filepath.Join(filepath.Dir(datadir), "testdata")

	// setup the local directory
	dir, err := ioutil.TempDir("", "kyaml-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	if !assert.NoError(t, copyutil.CopyDir(
		filepath.Join(datadir, "dataset1-localupdates"),
		filepath.Join(dir, "dataset1"))) {
		t.FailNow()
	}

	err = filters.Merge3{
		OriginalPath: filepath.Join(datadir, "dataset1"),
		UpdatedPath:  filepath.Join(datadir, "dataset1-remoteupdates"),
		DestPath:     filepath.Join(dir, "dataset1"),
	}.Merge()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	diffs, err := copyutil.Diff(
		filepath.Join(dir, "dataset1"),
		filepath.Join(datadir, "dataset1-expected"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Empty(t, diffs.List()) {
		t.FailNow()
	}
}
