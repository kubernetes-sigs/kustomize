// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusterr

import (
	"fmt"
	"testing"
)

const (
	filepath = "/path/to/whatever"
	expected = "YAML file [/path/to/whatever] encounters a format error.\n" +
		"error converting YAML to JSON: yaml: line 2: found character that cannot start any token\n"
)

func TestYamlFormatError_Error(t *testing.T) {
	testErr := YamlFormatError{
		Path:     filepath,
		ErrorMsg: "error converting YAML to JSON: yaml: line 2: found character that cannot start any token",
	}
	if testErr.Error() != expected {
		t.Errorf("Expected : %s\n, but found : %s\n", expected, testErr.Error())
	}
}

func TestErrorHandler(t *testing.T) {
	err := fmt.Errorf("error converting YAML to JSON: yaml: line 2: found character that cannot start any token")
	testErr := Handler(err, filepath)
	expectedErr := fmt.Errorf("format error message")
	fmtErr := Handler(expectedErr, filepath)
	if fmtErr.Error() != expectedErr.Error() {
		t.Errorf("Expected returning fmt.Error, but found %T", fmtErr)
	}
	if _, ok := testErr.(YamlFormatError); !ok {
		t.Errorf("Expected returning YamlFormatError, but found %T", testErr)
	}
	if testErr == nil || testErr.Error() != expected {
		t.Errorf("Expected : %s\n, but found : %s\n", expected, testErr.Error())
	}
}
