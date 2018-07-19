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

package error

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	filepath = "/path/to/" + constants.KustomizationFileName
	expected = "YAML file [/path/to/kustomization.yaml] encounters a format error.\n" +
		"error converting YAML to JSON: yaml: line 2: found character that cannot start any token\n"
	doc = `
	foo:
	- fiz
	- fu
	`
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
	f := foo{}
	err := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(doc))).Decode(&f)
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

//type foo struct
type foo struct{}
