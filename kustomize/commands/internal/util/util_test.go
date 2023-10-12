// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestConvertToMap(t *testing.T) {
	args := "a:b,c:\"d\",e:\"f:g\",g:h:k"
	expected := make(map[string]string)
	expected["a"] = "b"
	expected["c"] = "d"
	expected["e"] = "f:g"
	expected["g"] = "h:k"

	result, err := ConvertToMap(args, "annotation")
	require.NoError(t, err, "unexpected error")

	eq := reflect.DeepEqual(expected, result)
	require.True(t, eq, "Converted map does not match expected")
}

func TestConvertToMapError(t *testing.T) {
	args := "a:b,c:\"d\",:f:g"

	_, err := ConvertToMap(args, "annotation")
	require.Error(t, err, "expected error but did not receive one")
	require.Equal(t, "invalid annotation: ':f:g' (need k:v pair where v may be quoted)", err.Error(), "incorrect error")
}

func TestConvertSliceToMap(t *testing.T) {
	args := []string{"a:b", "c:\"d\"", "e:\"f:g\"", "g:h:k"}
	expected := make(map[string]string)
	expected["a"] = "b"
	expected["c"] = "d"
	expected["e"] = "f:g"
	expected["g"] = "h:k"

	result, err := ConvertSliceToMap(args, "annotation")
	require.NoError(t, err, "unexpected error")

	eq := reflect.DeepEqual(expected, result)
	require.True(t, eq, "Converted map does not match expected")
}

func TestGlobPatternsWithLoaderRemoteFile(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.Create("test.yml")
	httpPath := "https://example.com/example.yaml"
	ldr := fakeLoader{
		path: httpPath,
	}

	// test load remote file
	resources, err := GlobPatternsWithLoader(fSys, ldr, []string{httpPath}, false)
	require.NoError(t, err, "unexpected load error")
	require.Equal(t, 1, len(resources), "incorrect resources")
	require.Equal(t, httpPath, resources[0], "incorrect resources")

	// test load local and remote file
	resources, err = GlobPatternsWithLoader(fSys, ldr, []string{httpPath, "/test.yml"}, false)
	require.NoError(t, err, "unexpected load error")
	require.Equal(t, 2, len(resources), "incorrect resources")
	require.Equal(t, httpPath, resources[0], "incorrect resources")
	require.Equal(t, "/test.yml", resources[1], "incorrect resources")

	// test load invalid file
	invalidURL := "http://invalid"
	resources, err = GlobPatternsWithLoader(fSys, ldr, []string{invalidURL}, false)
	require.Error(t, err, "expected error but did not receive one")
	require.Equal(t, invalidURL+" has no match: "+invalidURL+" not exist", err.Error(), "unexpected load error")
	require.Equal(t, 0, len(resources), "incorrect resources")

	// test load unreachable remote file with skipped verification
	resources, err = GlobPatternsWithLoader(fSys, ldr, []string{invalidURL}, true)
	require.NoError(t, err, "unexpected load error")
	require.Equal(t, 1, len(resources), "incorrect resources")
	require.Equal(t, invalidURL, resources[0], "incorrect resources")
}

type fakeLoader struct {
	path string
}

func (l fakeLoader) Repo() string {
	return ""
}
func (l fakeLoader) Root() string {
	return ""
}
func (l fakeLoader) New(newRoot string) (ifc.Loader, error) {
	if newRoot == l.path {
		return nil, nil
	}
	return nil, fmt.Errorf("%s not exist", newRoot)
}
func (l fakeLoader) Load(location string) ([]byte, error) {
	return nil, nil
}
func (l fakeLoader) Cleanup() error {
	return nil
}
