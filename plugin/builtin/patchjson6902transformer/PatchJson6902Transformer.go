// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filters/patchjson6902"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	ldr          ifc.Loader
	decodedPatch jsonpatch.Patch
	Target       types.PatchTarget `json:"target,omitempty" yaml:"target,omitempty"`
	Path         string            `json:"path,omitempty" yaml:"path,omitempty"`
	JsonOp       string            `json:"jsonOp,omitempty" yaml:"jsonOp,omitempty"`

	YAMLSupport bool `json:"yamlSupport,omitempty" yaml:"yamlSupport,omitempty"`
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
	if !strings.Contains(string(c), "yamlSupport") {
		// If not explicitly denied,
		// activate kyaml-based transformation.
		p.YAMLSupport = true
	}
	if p.Target.Name == "" {
		return fmt.Errorf("must specify the target name")
	}
	if p.Path == "" && p.JsonOp == "" {
		return fmt.Errorf("empty file path and empty jsonOp")
	}
	if p.Path != "" {
		if p.JsonOp != "" {
			return fmt.Errorf("must specify a file path or jsonOp, not both")
		}
		rawOp, err := p.ldr.Load(p.Path)
		if err != nil {
			return err
		}
		p.JsonOp = string(rawOp)
		if p.JsonOp == "" {
			return fmt.Errorf("patch file '%s' empty seems to be empty", p.Path)
		}
	}
	if p.JsonOp[0] != '[' {
		// if it doesn't seem to be JSON, imagine
		// it is YAML, and convert to JSON.
		op, err := yaml.YAMLToJSON([]byte(p.JsonOp))
		if err != nil {
			return err
		}
		p.JsonOp = string(op)
	}
	p.decodedPatch, err = jsonpatch.DecodePatch([]byte(p.JsonOp))
	if err != nil {
		return errors.Wrapf(err, "decoding %s", p.JsonOp)
	}
	if len(p.decodedPatch) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; file=%s, JsonOp=%s", p.Path, p.JsonOp)
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
	if !p.YAMLSupport {
		rawObj, err := obj.MarshalJSON()
		if err != nil {
			return err
		}
		modifiedObj, err := p.decodedPatch.Apply(rawObj)
		if err != nil {
			return errors.Wrapf(
				err, "failed to apply json patch '%s'", p.JsonOp)
		}
		return obj.UnmarshalJSON(modifiedObj)
	} else {
		return filtersutil.ApplyToJSON(patchjson6902.Filter{
			Patch: p.JsonOp,
		}, obj)
	}
}
