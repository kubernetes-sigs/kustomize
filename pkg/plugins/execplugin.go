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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/yaml"
)

const (
	ArgsOneLiner = "argsOneLiner"
	ArgsFromFile = "argsFromFile"
	idAnnotation = "kustomize.config.k8s.io/id"
)

// ExecPlugin record the name and args of an executable
// It triggers the executable generator and transformer
type ExecPlugin struct {
	// name of the executable
	name string

	// Optional command line arguments to the executable
	// pulled from specially named fields in cfg.
	// This is for executables that don't want to parse YAML.
	args []string

	// Plugin configuration data.
	cfg []byte

	// resmap Factory to make resources
	rf *resmap.Factory

	// loader to load files
	ldr ifc.Loader
}

func NewExecPlugin(root string, id resid.ResId) *ExecPlugin {
	return &ExecPlugin{
		name: filepath.Join(root, pluginPath(id)),
	}
}

// isAvailable checks to see if the plugin is available
func (p *ExecPlugin) isAvailable() bool {
	f, err := os.Stat(p.name)
	if os.IsNotExist(err) {
		return false
	}
	return f.Mode()&0111 != 0000
}

func (p *ExecPlugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	p.rf = rf
	p.ldr = ldr

	var err error
	p.cfg, err = yaml.Marshal(k)
	if err != nil {
		return err
	}
	err = p.processOptionalArgsFields(k)
	if err != nil {
		return err
	}
	return nil
}

func (p *ExecPlugin) processOptionalArgsFields(k ifc.Kunstructured) error {
	args, err := k.GetFieldValue(ArgsOneLiner)
	if err == nil && args != "" {
		p.args = strings.Split(args, " ")
	}
	fileName, err := k.GetFieldValue(ArgsFromFile)
	if err == nil && fileName != "" {
		content, err := p.ldr.Load(fileName)
		if err != nil {
			return err
		}
		for _, x := range strings.Split(string(content), "\n") {
			x := strings.TrimLeft(x, " ")
			if x != "" {
				p.args = append(p.args, x)
			}
		}
	}
	return nil
}

func (p *ExecPlugin) writeConfig() (string, error) {
	tmpFile, err := ioutil.TempFile("", "kust-pipe")
	if err != nil {
		return "", err
	}
	syscall.Mkfifo(tmpFile.Name(), 0600)
	stdout, err := os.OpenFile(tmpFile.Name(), os.O_RDWR, 0600)
	if err != nil {
		return "", err
	}
	_, err = stdout.Write(p.cfg)
	if err != nil {
		return "", err
	}
	err = stdout.Close()
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func (p *ExecPlugin) Generate() (resmap.ResMap, error) {
	output, err := p.invokePlugin(nil)
	if err != nil {
		return nil, err
	}
	return p.rf.NewResMapFromBytes(output)
}

func (p *ExecPlugin) Transform(rm resmap.ResMap) error {
	// add ResIds as annotations to all objects so that we can add them back
	inputRM, err := p.getResMapWithIdAnnotation(rm)
	if err != nil {
		return err
	}

	// encode the ResMap so it can be fed to the plugin
	resources, err := inputRM.EncodeAsYaml()
	if err != nil {
		return err
	}

	// invoke the plugin with resources as the input
	output, err := p.invokePlugin(resources)
	if err != nil {
		return err
	}

	// update the original ResMap based on the output
	return p.updateResMapValues(output, rm)
}

// invokePlugin invokes the plugin
func (p *ExecPlugin) invokePlugin(input []byte) ([]byte, error) {
	args, err := p.getArgs()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(p.name, args...)
	cmd.Env = p.getEnv()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = os.Stderr
	if _, err := os.Stat(p.ldr.Root()); err == nil {
		cmd.Dir = p.ldr.Root()
	}
	return cmd.Output()
}

// The first arg is always the absolute path to a temporary file
// holding the YAML form of the plugin config.
func (p *ExecPlugin) getArgs() ([]string, error) {
	configFileName, err := p.writeConfig()
	if err != nil {
		return nil, err
	}
	return append([]string{configFileName}, p.args...), nil
}

func (p *ExecPlugin) getEnv() []string {
	env := os.Environ()
	env = append(env,
		"KUSTOMIZE_PLUGIN_CONFIG_STRING="+string(p.cfg),
		"KUSTOMIZE_PLUGIN_CONFIG_ROOT="+p.ldr.Root())
	return env
}

// Returns a new copy of the given ResMap with the ResIds annotated in each Resource
func (p *ExecPlugin) getResMapWithIdAnnotation(rm resmap.ResMap) (resmap.ResMap, error) {
	inputRM := rm.DeepCopy(p.rf.RF())
	for id, r := range inputRM {
		idString, err := yaml.Marshal(id)
		if err != nil {
			return nil, err
		}
		annotations := r.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{
				idAnnotation: string(idString),
			}
			r.SetAnnotations(annotations)
		}
		annotations[idAnnotation] = string(idString)
	}
	return inputRM, nil
}

/*
updateResMapValues updates the Resource value in the given ResMap
with the emitted Resource values in output.
*/
func (p *ExecPlugin) updateResMapValues(output []byte, rm resmap.ResMap) error {
	outputRM, err := p.rf.NewResMapFromBytes(output)
	if err != nil {
		return err
	}
	for _, r := range outputRM {
		// for each emitted Resource, find the matching Resource in the original ResMap
		// using its id
		annotations := r.GetAnnotations()
		idString, ok := annotations[idAnnotation]
		if !ok {
			return fmt.Errorf("the transformer %s should not remove annotation %s",
				p.name, idAnnotation)
		}
		id := resid.ResId{}
		err := yaml.Unmarshal([]byte(idString), &id)
		if err != nil {
			return err
		}
		res, ok := rm[id]
		if !ok {
			return fmt.Errorf("unable to find id %s in resource map", id.String())
		}
		// remove the annotation set by Kustomize to track the resource
		delete(annotations, idAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)

		// update the ResMap resource value with the transformed object
		res.Kunstructured = r.Kunstructured
	}
	return nil
}
