// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"

	"github.com/pkg/errors"
	"go.mozilla.org/sops/decrypt"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/yaml"
)

// Loads secrets from a sops-encoded file.
// See https://github.com/mozilla/sops
// Based on https://github.com/Agilicus/kustomize-sops
// and the sibling example SecretsFromDatabase.
type plugin struct {
	rf        *resmap.Factory
	ldr       ifc.Loader
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	File      string `json:"file,omitempty" file:"name,omitempty"`
	// List of keys to extract from secret map.
	Keys []string `json:"keys,omitempty" yaml:"keys,omitempty"`
}

//noinspection GoUnusedGlobalVariable
//nolint: golint
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	p.rf = rf
	p.ldr = ldr
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	secrets, err := p.loadSecretsViaSops()
	if err != nil {
		return nil, err
	}
	return p.makeK8sSecret(secrets)
}

func (p *plugin) loadSecretsViaSops() (secrets map[string]string, err error) {
	bytes, err := p.ldr.Load(p.File)
	if err != nil {
		return nil, errors.Wrapf(err, "trouble reading file %s", p.File)
	}
	var yamlMap []byte
	if isFakeEncryptedData(bytes) {
		yamlMap = []byte(fakeDecryptedData)
	} else {
		yamlMap, err = decrypt.Data(bytes, "yaml")
		if err != nil {
			return nil, errors.Wrapf(err, "decrypting content from %s", p.File)
		}
	}
	secrets = make(map[string]string)
	err = yaml.Unmarshal(yamlMap, &secrets)
	if err != nil {
		return nil, errors.Wrapf(
			err, "unmarshal failure from '%s'", string(yamlMap))
	}
	return
}

func (p *plugin) makeK8sSecret(
	secrets map[string]string) (resmap.ResMap, error) {
	args := types.SecretArgs{}
	args.Name = p.Name
	args.Namespace = p.Namespace
	for _, k := range p.Keys {
		if v, ok := secrets[k]; ok {
			args.LiteralSources = append(
				args.LiteralSources, k+"="+v)
		}
	}
	return p.rf.FromSecretArgs(p.ldr, nil, args)
}

// See test for justification of this hackery.
// The test is meant to just cover plugin behavior,
// and assume that sops works.  There's currently
// no way to inject a "mock" sops into this plugin.
const fakeDecryptedData = `
VEGETABLE: carrot
ROCKET: saturn-v
FRUIT: apple
CAR: dymaxion
`

func isFakeEncryptedData(bytes []byte) bool {
	return strings.Contains(string(bytes), "__ELIDED_FOR_KUSTOMIZE_TEST__")
}
