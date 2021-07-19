// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

type stringMarshal struct {
	Field string `yaml:",flow"`
}

type field struct {
	decodedPatch jsonpatch.Patch
	Name         string `json:"name" yaml:"name"`
	Path         string `json:"path,omitempty" yaml:"path,omitempty"`
	JsonOp       string `json:"jsonOp,omitempty" yaml:"jsonOp,omitempty"`
}

type plugin struct {
	ldr    ifc.Loader
	Target *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
	Fields []*field          `json:"fields,omitempty" yaml:"fields,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	h *resmap.PluginHelpers, c []byte) (err error) {
	p.ldr = h.Loader()
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	if p.Target.Name == "" {
		return fmt.Errorf("must specify the target name")
	}

	for _, field := range p.Fields {
		if lerr := p.load(field); lerr != nil {
			return lerr
		}
	}

	return err
}

func (p *plugin) Transform(m resmap.ResMap) error {
	id := resid.NewResIdWithNamespace(
		resid.Gvk{
			Group:   p.Target.Group,
			Version: p.Target.Version,
			Kind:    p.Target.Kind,
		},
		p.Target.Name,
		p.Target.Namespace,
	)
	obj, err := m.GetById(id)
	if err != nil {
		return err
	}
	if obj.GetKind() != "ConfigMap" {
		return fmt.Errorf("This patch cannot be applied to non configmap resources")
	}

	configMapData, gerr := obj.GetFieldValue("data")
	if gerr != nil {
		return gerr
	}
	configData := configMapData.(map[string]interface{})
	if configData == nil {
		return fmt.Errorf("Invalid data in configmap")
	}
	updatedData := make(map[string]string)

	for k,v := range configData {
		updatedData[k] = v.(string)
	}

	for _, field := range p.Fields {
		var kvalue interface{}
		var ok bool
		if kvalue, ok = updatedData[field.Name]; !ok {
			return fmt.Errorf("Non existent key %s in configmap", field.Name)
		}
		res, aerr := field.decodedPatch.Apply([]byte(kvalue.(string)))
		if aerr != nil {
			return fmt.Errorf("%s for key %s", aerr.Error(), field.Name)
		}
		var out bytes.Buffer
		if ierr := json.Indent(&out, res, "", "  "); ierr != nil {
			return ierr
		}

		updatedData[field.Name] = out.String()
	}

	obj.SetDataMap(updatedData)
	return nil
}

func (p *plugin) load(field *field) (err error) {
	if field.Path == "" && field.JsonOp == "" {
		return fmt.Errorf("%s: empty file path and empty jsonOp", field.Name)
	}
	if field.Path != "" {
		if field.JsonOp != "" {
			return fmt.Errorf("%s: must specify a file path or jsonOp, not both", field.Name)
		}
		rawOp, err := p.ldr.Load(field.Path)
		if err != nil {
			return err
		}
		field.JsonOp = string(rawOp)
		if field.JsonOp == "" {
			return fmt.Errorf("%s: patch file '%s' empty seems to be empty", field.Name, field.Path)
		}
	}
	if field.JsonOp[0] != '[' {
		// if it doesn't seem to be JSON, imagine
		// it is YAML, and convert to JSON.
		op, err := k8syaml.YAMLToJSON([]byte(field.JsonOp))
		if err != nil {
			return err
		}
		field.JsonOp = string(op)
	}
	field.decodedPatch, err = jsonpatch.DecodePatch([]byte(field.JsonOp))
	if err != nil {
		return errors.Wrapf(err, "decoding %s", field.JsonOp)
	}
	if len(field.decodedPatch) == 0 {
		return fmt.Errorf(
			"%s: patch appears to be empty; file=%s, JsonOp=%s", field.Name, field.Path, field.JsonOp)
	}
	return nil
}
