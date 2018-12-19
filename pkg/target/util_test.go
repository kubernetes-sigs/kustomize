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

package target

import (
	"testing"
)

func writeKustomizationFiles(th *KustTestHarness) {
	th.writeK("/app/base", `
namePrefix: a-
commonLabels:
  app: myApp
resources:
- deployment.yaml
- service.yaml
kind: Kustomization
apiVersion: v1beta1
`)
	th.writeF("/app/base/apply.yaml", `
namePrefix: a-
commonLabels:
  app: myApp
resources:
- deployment.yaml
- service.yaml
kind: Kustomization
apiVersion: b1
`)
	th.writeF("app/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
name: myService
spec:
selector:
backend: bungie
ports:
	- port: 7002
`)
}

func TestIsKustomizationFile(t *testing.T) {
	th := NewKustTestHarness(t, "/app/base")
	writeKustomizationFiles(th)
	if !IsKustomizationFile(th.ldr, "kustomization.yaml") {
		t.Errorf("kustomization.yaml in %s is expected to be a kustomization file", th.ldr.Root())
	}
	if !IsKustomizationFile(th.ldr, "apply.yaml") {
		t.Errorf("apply.yaml in %s is expected to be a kustomization file", th.ldr.Root())
	}
	if IsKustomizationFile(th.ldr, "service.yaml") {
		t.Errorf("service.yaml in %s is not kustomization file", th.ldr.Root())
	}
}

func TestLoadKustomizationFile(t *testing.T) {
	th := NewKustTestHarness(t, "/app/base")
	writeKustomizationFiles(th)
	_, err := loadKustFile(th.ldr, true)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	_, err = loadKustFile(th.ldr, false)
	if err == nil {
		t.Fatalf("Expected error.")
	}
	if err.Error() != "Found multiple files for Kustomization under /app/base" {
		t.Fatalf("Incorrect error.")
	}
}
