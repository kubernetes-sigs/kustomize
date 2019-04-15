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
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
)

const idAnnotation = "kustomize.config.k8s.io/id"

type ExecPlugin struct {
	name string
	rf   *resmap.Factory
	ldr  ifc.Loader
	cfg  []ifc.Kunstructured
}

func (p *ExecPlugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, name string, k []ifc.Kunstructured) error {
	dir := filepath.Join(pgmconfig.ConfigRoot(), "plugins")
	p.name = filepath.Join(dir, name)
	p.rf = rf
	p.ldr = ldr
	p.cfg = k
	return nil
}

func (p *ExecPlugin) Generate() (resmap.ResMap, error) {
	cmd := exec.Command(p.name)
	cmd.Env = os.Environ()
	content, err := marshalUnstructuredSlice(p.cfg)
	if err != nil {
		return nil, err
	}
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return p.rf.NewResMapFromBytes(output)
}

func (p *ExecPlugin) Transform(rm resmap.ResMap) error {
	config, err := marshalUnstructuredSlice(p.cfg)
	if err != nil {
		return err
	}
	inputRM := rm.DeepCopy(p.rf.RF())
	for id, r := range inputRM {
		annotations := r.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[idAnnotation] = id.String()
		r.SetAnnotations(annotations)
	}
	resources, err := inputRM.EncodeAsYaml()
	if err != nil {
		return err
	}

	var input []byte
	if len(config) > 0 {
		input = append(append(config, []byte(`---\n`)...), resources...)
	} else {
		input = resources
	}
	cmd := exec.Command(p.name)
	cmd.Env = os.Environ()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	outputRM, err := p.rf.NewResMapFromBytes(output)
	if err != nil {
		return err
	}
	for _, r := range outputRM {
		annotations := r.GetAnnotations()
		idString, ok := annotations[idAnnotation]
		if !ok {
			return fmt.Errorf("the transformer %s should not remove annotation %s", p.name, idAnnotation)
		}
		id, err := resid.NewResIdFromString(idString)
		if err != nil {
			return err
		}
		res, ok := rm[*id]
		if !ok {
			return fmt.Errorf("unable to find id %s in resource map", id.String())
		}
		delete(annotations, idAnnotation)
		if len(annotations) == 0 {
			r.SetAnnotations(nil)
		} else {
			r.SetAnnotations(annotations)
		}
		res.Kunstructured = r.Kunstructured
	}
	return nil
}
