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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

// Generator represents a Generator
type Generator struct {
	name  string
	input []byte
}

// NewGenerator returns a Generator
func NewGenerator(input []byte) (*Generator, error) {
	dir := os.Getenv(XDG_CONFIG_HOME)
	if len(dir) == 0 {
		return nil, fmt.Errorf("$%s is undefined", XDG_CONFIG_HOME)
	}

	name, err := getGroup(input)
	if err != nil {
		return nil, err
	}
	executable := filepath.Join(dir, PgmName, "plugins", name)
	return &Generator{
		name:  executable,
		input: input,
	}, nil
}

func (g *Generator) Run(dir string) ([]byte, error) {
	if !isPluginAvailable(g.name) {
		return nil, fmt.Errorf("Executable %s not found", g.name)
	}
	cmd := exec.Command(g.name, "generate", string(g.input))
	cmd.Dir = dir
	cmd.Env = os.Environ()
	return cmd.Output()
}

// getGroup returns the Group in `groupVersion` from a generator or transformer file
func getGroup(input []byte) (string, error) {
	var m map[string]interface{}
	err := yaml.Unmarshal(input, &m)
	if err != nil {
		return "", err
	}
	rawGV, ok := m["groupVersion"]
	if !ok {
		return "", fmt.Errorf("missing groupVersion in %s", string(input))
	}
	gv, ok := rawGV.(string)
	if !ok {
		return "", fmt.Errorf("can't convert %v to string", rawGV)
	}
	values := strings.Split(gv, "/")
	return values[0], nil
}

// isPluginAvailable checks if a plugin is available from
// $XDG_CONFIG_HOME/kustomize/plugins/
func isPluginAvailable(name string) bool {
	f, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	if f.Mode()&0111 != 0000 {
		return true
	}
	return false
}
