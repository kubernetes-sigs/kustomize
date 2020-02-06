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

// TestMerge3_Merge_path tests that if the same resource is specified multiple times
// with MergeOnPath, that the resources will be merged by the filepath name.
func TestMerge3_Merge_path(t *testing.T) {
	_, datadir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	datadir = filepath.Join(filepath.Dir(datadir), "testdata2")

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
		MergeOnPath:  true,
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

// TestMerge3_Merge_fail tests that if the same resource is defined multiple times
// that merge will fail
func TestMerge3_Merge_fail(t *testing.T) {
	_, datadir, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	datadir = filepath.Join(filepath.Dir(datadir), "testdata2")

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
	if !assert.Error(t, err) {
		t.FailNow()
	}
}
