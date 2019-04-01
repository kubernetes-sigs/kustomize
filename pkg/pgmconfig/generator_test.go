/*
Copyright 2019 The Kubernetes Authors.

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

package pgmconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	testcases := []struct {
		env    string
		input  string
		group  string
		errmsg string
	}{
		{
			input:  "anything",
			errmsg: "$XDG_CONFIG_HOME is undefined",
		},
		{
			env:    "test",
			input:  "anything",
			errmsg: "error unmarshaling JSON: json: cannot unmarshal string into Go value of type map[string]interface {}",
		},
		{
			env:    "test",
			input:  "apiVersion: v1",
			errmsg: "missing groupVersion in apiVersion: v1",
		},
		{
			env:   "test",
			input: "groupVersion: team.example.com/v1beta1",
			group: "team.example.com",
		},
	}
	for _, testcase := range testcases {
		if testcase.env != "" {
			os.Setenv(XDG_CONFIG_HOME, testcase.env)
		} else {
			os.Unsetenv(XDG_CONFIG_HOME)
		}
		g, err := NewGenerator([]byte(testcase.input))

		if testcase.errmsg == "" {
			if err != nil {
				t.Errorf("unpected error %v", err)
			}
			expected := filepath.Join(testcase.env, "kustomize", "plugins", testcase.group)
			if g.name != expected {
				t.Errorf("expected executable as %s, but got %s", expected, g.name)
			}
		} else {
			if err == nil {
				t.Errorf("expected error %s, but not happened", testcase.errmsg)
			}
			if err.Error() != testcase.errmsg {
				t.Errorf("expected %s, but got %v", testcase.errmsg, err)
			}
		}
	}
}

func TestGeneratorRun(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	os.Setenv(XDG_CONFIG_HOME, filepath.Join(dir, "testdata"))
	input := []byte(`
groupVersion: test.kustomize.k8s.io/v1beta
kind: ConfigMapG
generate:
  name: test
`)
	g, err := NewGenerator(input)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	output, err := g.Run("./")
	expect := `
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-configmap-test
data:
  username: admin
  password: secret
`
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if string(output) != expect {
		t.Errorf("expected %s, but got %s", expect, string(output))
	}
}
