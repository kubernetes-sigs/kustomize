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
	"os"
	"os/exec"
	"path/filepath"
	"sigs.k8s.io/kustomize/pkg/types"
	"strings"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

// ExecPlugin record the name and args of an executable
// It triggers the executable generator and transformer
type ExecPlugin struct {
	// name of the executable
	name string

	// one line of arguments for the executable
	argOneLiner string

	// relative file path to a file
	// Each line of this file is treated as one argument
	argsFromFile string

	// cfg hold the unstructured data which can be used
	// to configure the plugin
	cfg string

	// resmap Factory to make resources
	rf *resmap.Factory

	// loader to load files
	ldr ifc.Loader
}

func (p *ExecPlugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	dir := filepath.Join(pgmconfig.ConfigRoot(), "plugins")
	id := k.GetGvk()
	p.name = filepath.Join(dir, id.Group, id.Version, id.Kind)
	p.rf = rf
	p.ldr = ldr

	var err error
	data, err := yaml.Marshal(k)
	if err != nil {
		return err
	}
	p.cfg = string(data)

	p.argOneLiner, err = k.GetFieldValue("arg")
	if err != nil && !isNoFieldError(err) {
		return err
	}
	p.argsFromFile, err = k.GetFieldValue("file")
	if err != nil && !isNoFieldError(err) {
		return err
	}
	return nil
}

func (p *ExecPlugin) Generate() (resmap.ResMap, error) {
	args, err := p.getArgs()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(p.name, args...)
	cmd.Env = p.getEnv()
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return p.rf.NewResMapFromBytes(output)
}

func (p *ExecPlugin) Transform(rm resmap.ResMap) error {
	args, err := p.getArgs()
	if err != nil {
		return err
	}

	for id, r := range rm {
		content, err := yaml.Marshal(r.Kunstructured)
		if err != nil {
			return err
		}
		cmd := exec.Command(p.name, args...)
		cmd.Env = p.getEnv()
		cmd.Stdin = bytes.NewReader(content)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		tmpMap, err := p.rf.NewResMapFromBytes(output)
		if err != nil {
			return err
		}
		if len(tmpMap) != 1 {
			return fmt.Errorf("Unable to put two resources into one")
		}
		for _, v := range tmpMap {
			rm[id].Kunstructured = v.Kunstructured
		}
	}
	return nil
}

func (p *ExecPlugin) getArgs() ([]string, error) {
	args := strings.Split(p.argOneLiner, " ")
	if p.argsFromFile != "" {
		content, err := p.ldr.Load(p.argsFromFile)
		if err != nil {
			return nil, err
		}
		args = append(args, strings.Split(string(content), "\n")...)
	}
	return args, nil
}

func (p *ExecPlugin) getEnv() []string {
	env := os.Environ()
	env = append(env, "KUSTOMIZE_PLUGIN_CONFIG_STRING="+p.cfg)
	return env
}

func isNoFieldError(e error) bool {
	_, ok := e.(types.NoFieldError)
	if ok {
		return true
	}
	return false
}
