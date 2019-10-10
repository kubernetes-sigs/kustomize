// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	idAnnotation        = "kustomize.config.k8s.io/id"
	hashAnnotation      = "kustomize.config.k8s.io/needs-hash"
	behaviorAnnotation  = "kustomize.config.k8s.io/behavior"
	tmpConfigFilePrefix = "kust-plugin-config-"
)

// ExecPlugin record the name and args of an executable
// It triggers the executable generator and transformer
type ExecPlugin struct {
	// absolute path of the executable
	path string

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

func NewExecPlugin(p string) *ExecPlugin {
	return &ExecPlugin{path: p}
}

// isAvailable checks to see if the plugin is available
func (p *ExecPlugin) isAvailable() bool {
	f, err := os.Stat(p.path)
	if os.IsNotExist(err) {
		return false
	}
	return f.Mode()&0111 != 0000
}

func (p *ExecPlugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, config []byte) error {
	p.rf = rf
	p.ldr = ldr
	p.cfg = config
	return p.processOptionalArgsFields()
}

type argsConfig struct {
	ArgsOneLiner string `json:"argsOneLiner,omitempty" yaml:"argsOneLiner,omitempty"`
	ArgsFromFile string `json:"argsFromFile,omitempty" yaml:"argsFromFile,omitempty"`
}

func (p *ExecPlugin) processOptionalArgsFields() error {
	var c argsConfig
	yaml.Unmarshal(p.cfg, &c)
	if c.ArgsOneLiner != "" {
		p.args = strings.Split(c.ArgsOneLiner, " ")
	}
	if c.ArgsFromFile != "" {
		content, err := p.ldr.Load(c.ArgsFromFile)
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

func (p *ExecPlugin) Generate() (resmap.ResMap, error) {
	output, err := p.invokePlugin(nil)
	if err != nil {
		return nil, err
	}
	rm, err := p.rf.NewResMapFromBytes(output)
	if err != nil {
		return nil, err
	}
	return p.updateResourceOptions(rm)
}

func (p *ExecPlugin) Transform(rm resmap.ResMap) error {
	// add ResIds as annotations to all objects so that we can add them back
	inputRM, err := p.getResMapWithIdAnnotation(rm)
	if err != nil {
		return err
	}

	// encode the ResMap so it can be fed to the plugin
	resources, err := inputRM.AsYaml()
	if err != nil {
		return err
	}

	// invoke the plugin with resources as the input
	output, err := p.invokePlugin(resources)
	if err != nil {
		return fmt.Errorf("%v %s", err, string(output))
	}

	// update the original ResMap based on the output
	return p.updateResMapValues(output, rm)
}

// invokePlugin writes plugin config to a temp file, then
// passes the full temp file path as the first arg to a process
// running the plugin binary.  Process output is returned.
func (p *ExecPlugin) invokePlugin(input []byte) ([]byte, error) {
	f, err := ioutil.TempFile("", tmpConfigFilePrefix)
	if err != nil {
		return nil, errors.Wrap(
			err, "creating tmp plugin config file")
	}
	_, err = f.Write(p.cfg)
	if err != nil {
		return nil, errors.Wrap(
			err, "writing plugin config to "+f.Name())
	}
	err = f.Close()
	if err != nil {
		return nil, errors.Wrap(
			err, "closing plugin config file "+f.Name())
	}
	cmd := exec.Command(
		p.path, append([]string{f.Name()}, p.args...)...)
	cmd.Env = p.getEnv()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = os.Stderr
	if _, err := os.Stat(p.ldr.Root()); err == nil {
		cmd.Dir = p.ldr.Root()
	}
	result, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrapf(
			err, "failure in plugin configured via %s; %v",
			f.Name(), err.Error())
	}
	return result, os.Remove(f.Name())
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
	inputRM := rm.DeepCopy()
	for _, r := range inputRM.Resources() {
		idString, err := yaml.Marshal(r.CurId())
		if err != nil {
			return nil, err
		}
		annotations := r.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[idAnnotation] = string(idString)
		r.SetAnnotations(annotations)
	}
	return inputRM, nil
}

// updateResMapValues updates the Resource value in the given ResMap
// with the emitted Resource values in output.
func (p *ExecPlugin) updateResMapValues(output []byte, rm resmap.ResMap) error {
	outputRM, err := p.rf.NewResMapFromBytes(output)
	if err != nil {
		return err
	}
	for _, r := range outputRM.Resources() {
		// for each emitted Resource, find the matching Resource in the original ResMap
		// using its id
		annotations := r.GetAnnotations()
		idString, ok := annotations[idAnnotation]
		if !ok {
			return fmt.Errorf("the transformer %s should not remove annotation %s",
				p.path, idAnnotation)
		}
		id := resid.ResId{}
		err := yaml.Unmarshal([]byte(idString), &id)
		if err != nil {
			return err
		}
		res, err := rm.GetByCurrentId(id)
		if err != nil {
			return fmt.Errorf("unable to find unique match to %s", id.String())
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

// updateResourceOptions updates the generator options for each resource in the
// given ResMap based on plugin provided annotations.
func (p *ExecPlugin) updateResourceOptions(rm resmap.ResMap) (resmap.ResMap, error) {
	for _, r := range rm.Resources() {
		// Disable name hashing by default and require plugin to explicitly
		// request it for each resource.
		annotations := r.GetAnnotations()
		behavior := annotations[behaviorAnnotation]
		var needsHash bool
		if val, ok := annotations[hashAnnotation]; ok {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("the annotation %q contains an invalid value (%q)", hashAnnotation, val)
			}
			needsHash = b
		}
		delete(annotations, hashAnnotation)
		delete(annotations, behaviorAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)
		r.SetOptions(types.NewGenArgs(
			&types.GeneratorArgs{Behavior: behavior},
			&types.GeneratorOptions{DisableNameSuffixHash: !needsHash}))
	}
	return rm, nil
}
