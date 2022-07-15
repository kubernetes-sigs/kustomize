// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/kyaml/errors"
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
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestConvertToMapError(t *testing.T) {
	args := "a:b,c:\"d\",:f:g"

	_, err := ConvertToMap(args, "annotation")
	require.EqualError(t, err, "invalid annotation: ':f:g' (need k:v pair where v may be quoted)")
}

func TestConvertSliceToMap(t *testing.T) {
	args := []string{"a:b", "c:\"d\"", "e:\"f:g\"", "g:h:k"}
	expected := make(map[string]string)
	expected["a"] = "b"
	expected["c"] = "d"
	expected["e"] = "f:g"
	expected["g"] = "h:k"

	result, err := ConvertSliceToMap(args, "annotation")
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestGlobPatternsWithLoaderRemoteFile(t *testing.T) {
	req := require.New(t)

	fSys := filesys.MakeFsInMemory()
	_, err := fSys.Create("test.yml")
	req.NoError(err)

	httpPath := "https://example.com/example.yaml"
	ldr := fakeRemoteLoader{
		files: map[string]struct{}{httpPath: {}},
	}
	expected := [2]string{httpPath, "/test.yml"}

	// test load remote file
	resources, err := globPatternsWithLoader(fSys, ldr, []string{httpPath})
	req.NoError(err)
	req.Equal(expected[:1], resources)

	// test load local and remote file
	resources, err = globPatternsWithLoader(fSys, ldr, []string{httpPath, "/test.yml"})
	req.NoError(err)
	req.Equal(expected[:], resources)

	// test load invalid file
	resources, err = globPatternsWithLoader(fSys, ldr, []string{"http://invalid"})
	req.NoError(err)
	req.Empty(resources)
}

type fakeRemoteLoader struct {
	dirs  map[string]struct{}
	files map[string]struct{}
}

func (l fakeRemoteLoader) Root() string {
	return ""
}
func (l fakeRemoteLoader) New(newRoot string) (ifc.Loader, error) {
	if _, ok := l.dirs[newRoot]; ok {
		return nil, nil
	}
	return nil, errors.WrapPrefixf(
		errors.Errorf("does not exist"),
		fmt.Sprintf("'%s'", newRoot))
}
func (l fakeRemoteLoader) Load(location string) ([]byte, error) {
	if _, ok := l.files[location]; ok {
		return nil, nil
	}
	return nil, errors.WrapPrefixf(
		errors.Errorf("does not exist"),
		fmt.Sprintf("'%s'", location))
}
func (l fakeRemoteLoader) Cleanup() error {
	return nil
}
