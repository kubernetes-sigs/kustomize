// +build plugin

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

type plugin struct {
	name string
	arg  string
	file string
	rf   *resmap.Factory
	ldr  ifc.Loader
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, k ifc.Kunstructured) error {
	dir := filepath.Join(pgmconfig.ConfigRoot(), "plugins")
	id := k.GetGvk()
	p.name = filepath.Join(dir, id.Group, id.Version, id.Kind)
	p.rf = rf
	p.ldr = ldr

	var err error
	p.arg, err = k.GetFieldValue("arg")
	if err != nil && !strings.Contains(err.Error(), "no field named") {
		return err
	}
	p.file, err = k.GetFieldValue("file")
	if err != nil && !strings.Contains(err.Error(), "no field named") {
		return nil
	}
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.run(nil)
}

func (p *plugin) Transform(rm resmap.ResMap) error {
	_, err := p.run(rm)
	if err != nil {
		return err
	}
	return nil
}

func (p *plugin) run(rm resmap.ResMap) (resmap.ResMap, error) {
	args := strings.Split(p.arg, " ")
	if p.file != "" {
		content, err := p.ldr.Load(p.file)
		if err != nil {
			return nil, err
		}
		args = append(args, strings.Split(string(content), "\n")...)
	}

	// For generators
	if rm == nil {
		cmd := exec.Command(p.name, args...)
		cmd.Env = os.Environ()
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return p.rf.NewResMapFromBytes(output)
	}
	// For transformers
	for id, r := range rm {
		content, err := yaml.Marshal(r.Kunstructured)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(p.name, args...)
		cmd.Env = os.Environ()
		cmd.Stdin = bytes.NewReader(content)
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		tmpMap, err := p.rf.NewResMapFromBytes(output)
		if err != nil {
			return nil, err
		}
		if len(tmpMap) != 1 {
			return nil, fmt.Errorf("Unable to put two resources into one")
		}
		for _, v := range tmpMap {
			rm[id].Kunstructured = v.Kunstructured
		}
	}
	return rm, nil
}
