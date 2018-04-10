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
	"fmt"
	"strings"
	"testing"
)

func TestManifestError_Error(t *testing.T) {
	filepath := "/path/to/kustomize.yaml"
	errorMsg := "Manifest not found"

	me := ManifestError{ManifestFilepath: filepath, ErrorMsg: errorMsg}

	if !strings.Contains(me.Error(), filepath) {
		t.Errorf("Incorrect ManifestError.Error() message \n")
		t.Errorf("Expected filepath %s, but unfound\n", filepath)
	}

	if !strings.Contains(me.Error(), errorMsg) {
		t.Errorf("Incorrect ManifestError.Error() message \n")
		t.Errorf("Expected errorMsg %s, but unfound\n", errorMsg)
	}

}

func TestManifestErrors_Error(t *testing.T) {
	filepath := "/path/to/kustomize"
	me := ManifestError{ManifestFilepath: filepath, ErrorMsg: "Manifest not found"}
	ce := ConfigmapError{ManifestFilepath: filepath, ErrorMsg: "can't find configmap name"}
	pe := PatchError{ManifestFilepath: filepath, PatchFilepath: filepath, ErrorMsg: "can't find patch file"}
	re := ResourceError{ManifestFilepath: filepath, ResourceFilepath: filepath, ErrorMsg: "can't find resource file"}
	se := SecretError{ManifestFilepath: filepath, ErrorMsg: "can't find secret name"}
	mes := ManifestErrors{merrors: []error{me, ce, pe, re, se}}
	expectedErrorMsg := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", me.Error(), ce.Error(), pe.Error(), re.Error(), se.Error())
	if mes.Error() != expectedErrorMsg {
		t.Errorf("Incorrect ManifestErrors.Error() message\n")
		t.Errorf(" Expected: %s\n", expectedErrorMsg)
		t.Errorf(" Got: %s\n", mes.Error())
	}
}

func TestManifestErrors_Get(t *testing.T) {
	ce := ConfigmapError{ManifestFilepath: "manifest/filepath", ErrorMsg: "can't find configmap name"}
	mes := ManifestErrors{merrors: []error{ce}}
	if len(mes.Get()) != 1 {
		t.Errorf("Incorrect ManifestErrors.Get()\n")
		t.Errorf(" Expected: %v\n", []error{ce})
		t.Errorf(" Got: %s\n", mes.Get())
	}
}

func TestManifestErrors_Append(t *testing.T) {
	ce := ConfigmapError{ManifestFilepath: "manifest/filepath", ErrorMsg: "can't find configmap name"}
	pe := PatchError{ManifestFilepath: "manifest/filepath", PatchFilepath: "patch/path", ErrorMsg: "can't find patch file"}
	mes := ManifestErrors{merrors: []error{ce}}
	mes.Append(pe)
	if len(mes.Get()) != 2 {
		t.Errorf("Incorrect ManifestErrors.Append()\n")
		t.Errorf(" Expected: %d error\n%v/n", 2, []error{ce, pe})
		t.Errorf(" Got: %d error\n%v\n", len(mes.Get()), mes.Get())
	}
}

func TestManifestErrors_BatchAppend(t *testing.T) {
	ce := ConfigmapError{ManifestFilepath: "manifest/filepath", ErrorMsg: "can't find configmap name"}
	pe := PatchError{ManifestFilepath: "manifest/filepath", PatchFilepath: "patch/path", ErrorMsg: "can't find patch file"}
	mes := ManifestErrors{merrors: []error{ce}}
	me := ManifestErrors{merrors: []error{pe}}
	mes.BatchAppend(me)
	if len(mes.Get()) != 2 {
		t.Errorf("Incorrect ManifestErrors.Append()\n")
		t.Errorf(" Expected: %d error\n%v/n", 2, []error{ce, pe})
		t.Errorf(" Got: %d error\n%v\n", len(mes.Get()), mes.Get())
	}
}
