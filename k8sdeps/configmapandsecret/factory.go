// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package configmapandsecret

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/kustomize/k8sdeps/kv"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

// Factory makes ConfigMaps and Secrets.
type Factory struct {
	ldr     ifc.Loader
	options *types.GeneratorOptions
}

// NewFactory returns a new Factory.
func NewFactory(
	l ifc.Loader, o *types.GeneratorOptions) *Factory {
	return &Factory{ldr: l, options: o}
}

func (f *Factory) loadKvPairs(
	args types.GeneratorArgs) (all []kv.Pair, err error) {
	pairs, err := f.keyValuesFromEnvFiles(args.EnvSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"env source files: %v",
			args.EnvSources))
	}
	all = append(all, pairs...)

	pairs, err = keyValuesFromLiteralSources(args.LiteralSources)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(
			"literal sources %v", args.LiteralSources))
	}
	all = append(all, pairs...)

	pairs, err = f.keyValuesFromFileSources(args.FileSources)
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

func (f *Factory) keyValuesFromFileSources(sources []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, s := range sources {
		k, fPath, err := kv.ParseFileSource(s)
		if err != nil {
			return nil, err
		}
		content, err := f.ldr.Load(fPath)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv.Pair{Key: k, Value: string(content)})
	}
	return kvs, nil
}

func (f *Factory) keyValuesFromEnvFiles(paths []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, path := range paths {
		content, err := f.ldr.Load(path)
		if err != nil {
			return nil, err
		}
		more, err := kv.KeyValuesFromLines(content)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, more...)
	}
	return kvs, nil
}
