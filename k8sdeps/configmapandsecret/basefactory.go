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

package configmapandsecret

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/kustomize/k8sdeps/kv"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

// baseFactory holds code shared by Factory and SecretFactory.
type baseFactory struct {
	ldr     ifc.Loader
	options *types.GeneratorOptions
	reg     plugin.Registry
}

func (bf baseFactory) loadKvPairs(
	args types.GeneratorArgs) (all []kv.Pair, err error) {
	pairs, err := bf.keyValuesFromPlugins(args.KVSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"plugins: %s",
			args.EnvSource))
	}
	all = append(all, pairs...)

	pairs, err = bf.keyValuesFromEnvFile(args.EnvSource)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"env source file: %s",
			args.EnvSource))
	}
	all = append(all, pairs...)

	pairs, err = keyValuesFromLiteralSources(args.LiteralSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"literal sources %v", args.LiteralSources))
	}
	all = append(all, pairs...)

	pairs, err = bf.keyValuesFromFileSources(args.FileSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"file sources: %v", args.FileSources))
	}
	return append(all, pairs...), nil
}

const keyExistsErrorMsg = "cannot add key %s, another key by that name already exists: %v"

func errIfInvalidKey(keyName string) error {
	if errs := validation.IsConfigMapKey(keyName); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid key name: %s",
			keyName, strings.Join(errs, ";"))
	}
	return nil
}

func keyValuesFromLiteralSources(sources []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, s := range sources {
		k, v, err := kv.ParseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv.Pair{Key: k, Value: v})
	}
	return kvs, nil
}

func (bf baseFactory) keyValuesFromPlugins(sources []types.KVSource) ([]kv.Pair, error) {
	var allKvs []kv.Pair
	for _, s := range sources {
		plug, err := bf.reg.Load(s.PluginType, s.Name)
		if err != nil {
			return nil, err
		}
		kvs, err := plug.Get(bf.reg.Root(), s.Args)
		if err != nil {
			return nil, err
		}
		allKvs = append(allKvs, kvs...)
	}
	return allKvs, nil
}

func (bf baseFactory) keyValuesFromFileSources(sources []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, s := range sources {
		k, fPath, err := kv.ParseFileSource(s)
		if err != nil {
			return nil, err
		}
		content, err := bf.ldr.Load(fPath)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv.Pair{Key: k, Value: string(content)})
	}
	return kvs, nil
}

func (bf baseFactory) keyValuesFromEnvFile(path string) ([]kv.Pair, error) {
	if path == "" {
		return nil, nil
	}
	content, err := bf.ldr.Load(path)
	if err != nil {
		return nil, err
	}
	return kv.KeyValuesFromLines(content)
}
