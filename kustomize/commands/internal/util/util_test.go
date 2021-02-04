package util

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
)

func TestConvertToMap(t *testing.T) {
	args := "a:b,c:\"d\",e:\"f:g\",g:h:k"
	expected := make(map[string]string)
	expected["a"] = "b"
	expected["c"] = "d"
	expected["e"] = "f:g"
	expected["g"] = "h:k"

	result, err := ConvertToMap(args, "annotation")
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}

	eq := reflect.DeepEqual(expected, result)
	if !eq {
		t.Errorf("Converted map does not match expected, expected: %v, result: %v\n", expected, result)
	}
}

func TestConvertToMapError(t *testing.T) {
	args := "a:b,c:\"d\",:f:g"

	_, err := ConvertToMap(args, "annotation")
	if err == nil {
		t.Errorf("expected an error")
	}
	if err.Error() != "invalid annotation: ':f:g' (need k:v pair where v may be quoted)" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestConvertSliceToMap(t *testing.T) {
	args := []string{"a:b", "c:\"d\"", "e:\"f:g\"", "g:h:k"}
	expected := make(map[string]string)
	expected["a"] = "b"
	expected["c"] = "d"
	expected["e"] = "f:g"
	expected["g"] = "h:k"

	result, err := ConvertSliceToMap(args, "annotation")
	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}

	eq := reflect.DeepEqual(expected, result)
	if !eq {
		t.Errorf("Converted map does not match expected, expected: %v, result: %v\n", expected, result)
	}
}

func TestGlobPatternsWithLoaderRemoteFile(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	fSys.Create("test.yml")
	httpPath := "https://example.com/example.yaml"
	ldr := fakeLoader{
		path: httpPath,
	}

	// test load remote file
	resources, err := GlobPatternsWithLoader(fSys, ldr, []string{httpPath})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(resources) != 1 || resources[0] != httpPath {
		t.Fatalf("incorrect resources: %v", resources)
	}

	// test load local and remote file
	resources, err = GlobPatternsWithLoader(fSys, ldr, []string{httpPath, "/test.yml"})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(resources) != 2 || resources[0] != httpPath || resources[1] != "/test.yml" {
		t.Fatalf("incorrect resources: %v", resources)
	}

	// test load invalid file
	resources, err = GlobPatternsWithLoader(fSys, ldr, []string{"http://invalid"})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(resources) > 0 {
		t.Fatalf("incorrect resources: %v", resources)
	}
}

type fakeLoader struct {
	path string
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
