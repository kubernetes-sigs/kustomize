package util

import (
	"reflect"
	"testing"
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
