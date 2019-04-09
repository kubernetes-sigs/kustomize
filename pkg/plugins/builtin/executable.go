// +build plugin

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

type plugin struct {
	name  string
	input string
	rf    *resmap.Factory
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	dir := filepath.Join(pgmconfig.ConfigRoot(), "plugins")
	id := k.GetGvk()
	p.name = filepath.Join(dir, id.Group, id.Version, id.Kind)
	content, err := yaml.Marshal(k)
	if err != nil {
		return err
	}
	p.input = string(content)
	p.rf = rf
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.run(nil)
}

func (p *plugin) Transformer(rm resmap.ResMap) error {
	result, err := p.run(rm)
	if err != nil {
		return err
	}
	for id := range rm {
		delete(rm, id)
	}
	for id, r := range result {
		rm[id] = r
	}
	return nil
}

func (p *plugin) run(rm resmap.ResMap) (resmap.ResMap, error) {
	cmd := exec.Command(p.name, p.input)
	cmd.Env = os.Environ()
	if rm != nil {
		content, err := rm.EncodeAsYaml()
		if err != nil {
			return nil, err
		}
		cmd.Stdin = bytes.NewReader(content)
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return p.rf.NewResMapFromBytes(output)
}
