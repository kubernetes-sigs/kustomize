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
)

// Transformer represents a Transformer
type Transformer struct {
	name  string
	input []byte
}

// NewTransformer returns a Transformer
// TODO: Add list of resources from ResMap as input or Run argument of the transformer
func NewTransformer(input []byte) (*Transformer, error) {
	name, err := getGroup(input)
	if err != nil {
		return nil, err
	}
	executable := filepath.Join(XDG_CONFIG_HOME, PgmName, "plugins", name)
	if !isPluginAvailable(executable) {
		return nil, fmt.Errorf("Executable %s not found", name)
	}
	return &Transformer{
		name:  name,
		input: input,
	}, nil
}

func (g *Transformer) Run(dir string) ([]byte, error) {
	cmd := exec.Command(g.name, "transform", string(g.input))
	cmd.Dir = dir
	cmd.Env = os.Environ()
	return cmd.Output()
}
