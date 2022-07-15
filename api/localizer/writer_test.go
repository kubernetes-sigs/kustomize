// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package localizer_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	lclzr "sigs.k8s.io/kustomize/api/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type testFile struct {
	path    string
	content string
}

func getKustSetup() ([]string, []testFile) {
	dirs := []string{
		"/newDir",
		"/scope/newDir",
	}
	files := []testFile{
		{
			path: "/scope/target/kustomization.yaml",
			content: `
resources:
- deployment.yaml
- ../target/.././target/newDir/secret.yaml
- https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/service.yaml`,
		},
		{
			path:    "/scope/target/deployment.yaml",
			content: "deployment configuration",
		},
		{
			path:    "/scope/target/newDir/secret.yaml",
			content: "secret configuration",
		},
	}
	return dirs, files
}

func makeFakeFsys(require *require.Assertions, dirs []string, files []testFile) filesys.FileSystem {
	fSys := filesys.MakeFsInMemory()
	for _, d := range dirs {
		err := fSys.MkdirAll(d)
		require.NoError(err)
	}
	for _, f := range files {
		err := fSys.WriteFile(f.path, []byte(f.content))
		require.NoError(err)
	}
	return fSys
}

func checkValidWrite(require *require.Assertions, actualPath string, actualErr error,
	rootDst string, expectedPath string, expectedContent string, fSys filesys.FileSystem) {
	require.NoError(actualErr)
	require.Equal(expectedPath, actualPath)

	actualContent, readErr := fSys.ReadFile(filepath.Join(rootDst, expectedPath))
	require.NoError(readErr)
	require.Equal([]byte(expectedContent), actualContent)
}

func TestWriterNoScope(t *testing.T) {
	require := require.New(t)
	dirs, files := getKustSetup()
	fSys := makeFakeFsys(require, dirs, files)

	const newDir = "/scope/newDir"
	wr, err := lclzr.NewWriter("/scope/target", "/scope/target",
		newDir, fSys)
	require.NoError(err)

	const remoteContent = "service configuration"
	remotePath, remoteErr := wr.Write(
		"https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/service.yaml",
		[]byte(remoteContent))
	checkValidWrite(require, remotePath, remoteErr, newDir,
		"localized-files/raw.githubusercontent.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/simple/service.yaml",
		remoteContent, fSys)

	const subDirContent = "secret configuration"
	subDirPath, subDirErr := wr.Write("../target/.././target/newDir/secret.yaml",
		[]byte(subDirContent))
	checkValidWrite(require, subDirPath, subDirErr, newDir, "newDir/secret.yaml",
		subDirContent, fSys)
}

func TestWriterNewDirInsideTarget(t *testing.T) {
	require := require.New(t)
	dirs, files := getKustSetup()
	fSys := makeFakeFsys(require, dirs, files)

	wr, err := lclzr.NewWriter("/scope/target", "/scope", "/scope/target/newDir", fSys)
	require.NoError(err)

	const locContent = "deployment configuration"
	locPath, locErr := wr.Write("deployment.yaml", []byte(locContent))
	checkValidWrite(require, locPath, locErr, "scope/target/newDir/target",
		"deployment.yaml", locContent, fSys)

	_, newDirErr := wr.Write("newDir/target/deployment.yaml", []byte(locContent))
	require.Error(newDirErr)
}
