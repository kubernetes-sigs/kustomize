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

package plugins

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"plugin"
	kplugin "sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

type Configurable interface {
	Config(ldr ifc.Loader, rf *resmap.Factory, name string, k []ifc.Kunstructured) error
}

func goPluginFileName(dir, name string) string {
	return execPluginFileName(dir, name) + ".so"
}

func execPluginFileName(dir, name string) string {
	return filepath.Join(dir, name)
}

// isExecAvailable checks if an executable is available
func isExecAvailable(name string) bool {
	f, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return f.Mode()&0111 != 0000
}

func loadAndConfigurePlugin(
	dir string, name string,
	ldr ifc.Loader,
	rf *resmap.Factory, res []ifc.Kunstructured) (Configurable, error) {
	var fileName string
	var c Configurable

	exec := execPluginFileName(dir, name)
	if isExecAvailable(exec) {
		c = &ExecPlugin{}
	} else {
		fileName = goPluginFileName(dir, name)

		goPlugin, err := plugin.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "plugin %s fails to load", fileName)
		}
		symbol, err := goPlugin.Lookup(kplugin.PluginSymbol)
		if err != nil {
			return nil, errors.Wrapf(
				err, "plugin %s doesn't have symbol %s",
				fileName, kplugin.PluginSymbol)
		}
		var ok bool
		c, ok = symbol.(Configurable)
		if !ok {
			return nil, fmt.Errorf("plugin %s not configurable", fileName)
		}
	}
	err := c.Config(ldr, rf, name, res)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s fails configuration", fileName)
	}
	return c, nil
}

func getGroupedConfigs(rm resmap.ResMap) map[string][]ifc.Kunstructured {
	result := make(map[string][]ifc.Kunstructured)
	for id, res := range rm {
		g := id.Gvk().Group
		_, ok := result[g]
		if ok {
			result[g] = append(result[g], res.Kunstructured)
		} else {
			result[g] = []ifc.Kunstructured{res.Kunstructured}
		}
	}
	return result
}

func marshalUnstructuredSlice(a []ifc.Kunstructured) ([]byte, error) {
	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, u := range a {
		out, err := yaml.Marshal(u.Map())
		if err != nil {
			return nil, err
		}
		if firstObj {
			firstObj = false
		} else {
			_, err = buf.WriteString("---\n")
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.Write(out)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}
