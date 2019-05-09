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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/plugins"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestPluginDir(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildExecPlugin(
		"someteam.example.com", "v1", "PrintWorkDir")

	base, err := os.Getwd()
	if err != nil {
		t.Fatalf("err %v", err)
	}
	dir, err := ioutil.TempDir(base, "kustomize-")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	defer os.RemoveAll(dir)

	fSys := fs.MakeRealFS()
	err = fSys.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte(`
generators:
- config.yaml
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}
	err = fSys.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
apiVersion: someteam.example.com/v1
kind: PrintWorkDir
metadata:
  name: some-random-name
`))
	if err != nil {
		t.Fatalf("err %v", err)
	}

	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, dir, fSys)
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	rf := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryWithGeneratorArgs(
			&types.GeneratorMetaArgs{PluginConfig: plugin.ActivePluginConfig()})))

	pl := plugins.NewLoader(rf, plugins.NewExternalPluginLoader(plugin.ActivePluginConfig(), rf))
	tg, err := target.NewKustTarget(ldr, rf, transformer.NewFactoryImpl(), pl)
	if err != nil {
		t.Fatalf("err %v", err)
	}

	m, err := tg.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	th := kusttest_test.NewKustTestHarness(t, ".")
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: WorkDir
metadata:
  name: `+dir+`
spec:
  path: `+dir+`
`)
}
