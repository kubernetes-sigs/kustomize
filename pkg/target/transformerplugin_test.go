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

package target_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fmt"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/types"
)

func writeDeployment(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - name: whatever
        image: whatever
`)
}

func writeStringPrefixer(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: strings.microwoosh.com/v1
kind: StringPrefixer
metadata:
  name: myStringPrefixer
prefix: apple-
`)
}

func writeDatePrefixer(th *KustTestHarness, path string) {
	th.writeF(path, `
apiVersion: team.dater.com/v1
kind: DatePrefixer
metadata:
  name: myDatePrefixer
`)
}

func buildGoPlugins(dir, filename string) error {
	commands := []string{
		"build",
		"-buildmode",
		"plugin",
		"-tags=plugin",
		"-o",
		filename + ".so",
		filename + ".go",
	}
	goBin := filepath.Join(os.Getenv("GOROOT"), "bin", "go")
	if _, err := os.Stat(goBin); err != nil {
		return fmt.Errorf("go binary not found %s", goBin)
	}
	cmd := exec.Command(goBin, commands...)
	cmd.Env = os.Environ()
	cmd.Dir = filepath.Join(dir, "kustomize", "plugins")

	return cmd.Run()
}

func TestOrderedTransformers(t *testing.T) {
	dir, err := filepath.Abs("../../..")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	os.Setenv(pgmconfig.XDG_CONFIG_HOME, dir)
	defer os.Unsetenv(pgmconfig.XDG_CONFIG_HOME)

	err = buildGoPlugins(dir, "StringPrefixer")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	err = buildGoPlugins(dir, "DatePrefixer")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	th := NewKustTestHarnessWithPluginConfig(
		t, "/app", types.PluginConfig{GoEnabled: true})
	th.writeK("/app", `
resources:
- deployment.yaml
transformers:
- stringPrefixer.yaml
`)
	writeDeployment(th, "/app/deployment.yaml")
	writeStringPrefixer(th, "/app/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/datePrefixer.yaml")
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apple-myDeployment
spec:
  template:
    metadata:
      labels:
        backend: awesome
    spec:
      containers:
      - image: whatever
        name: whatever
`)
}

func xTestTransformedTransformers(t *testing.T) {
	th := NewKustTestHarnessWithPluginConfig(
		t, "/app/overlay", types.PluginConfig{GoEnabled: true})

	th.writeK("/app/base", `
resources:
- stringPrefixer.yaml
transformers:
- datePrefixer.yaml
`)
	writeStringPrefixer(th, "/app/base/stringPrefixer.yaml")
	writeDatePrefixer(th, "/app/base/datePrefixer.yaml")

	th.writeK("/app/overlay", `
resources:
- deployment.yaml
transformers:
- ../base
`)
	writeDeployment(th, "/app/overlay/deployment.yaml")

	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, `
HEY
`)
}
