// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resmap

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/v3/internal/kusterr"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

// Factory makes instances of ResMap.
type Factory struct {
	resF *resource.Factory
	tf   PatchFactory
}

// NewFactory returns a new resmap.Factory.
func NewFactory(rf *resource.Factory, tf PatchFactory) *Factory {
	return &Factory{resF: rf, tf: tf}
}

// RF returns a resource.Factory.
func (rmF *Factory) RF() *resource.Factory {
	return rmF.resF
}

func New() ResMap {
	return newOne()
}

// FromResource returns a ResMap with one entry.
func (rmF *Factory) FromResource(res *resource.Resource) ResMap {
	m, err := newResMapFromResourceSlice([]*resource.Resource{res})
	if err != nil {
		panic(err)
	}
	return m
}

// FromFile returns a ResMap given a resource path.
func (rmF *Factory) FromFile(
	loader ifc.Loader, path string) (ResMap, error) {
	content, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	m, err := rmF.NewResMapFromBytes(content)
	if err != nil {
		return nil, kusterr.Handler(err, path)
	}
	return m, nil
}

// NewResMapFromBytes decodes a list of objects in byte array format.
func (rmF *Factory) NewResMapFromBytes(b []byte) (ResMap, error) {
	resources, err := rmF.resF.SliceFromBytes(b)
	if err != nil {
		return nil, err
	}
	return newResMapFromResourceSlice(resources)
}

// NewResMapFromConfigMapArgs returns a Resource slice given
// a configmap metadata slice from kustomization file.
func (rmF *Factory) NewResMapFromConfigMapArgs(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	argList []types.ConfigMapArgs) (ResMap, error) {
	var resources []*resource.Resource
	for _, args := range argList {
		res, err := rmF.resF.MakeConfigMap(ldr, options, &args)
		if err != nil {
			return nil, errors.Wrap(err, "NewResMapFromConfigMapArgs")
		}
		resources = append(resources, res)
	}
	return newResMapFromResourceSlice(resources)
}

func (rmF *Factory) FromConfigMapArgs(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args types.ConfigMapArgs) (ResMap, error) {
	res, err := rmF.resF.MakeConfigMap(ldr, options, &args)
	if err != nil {
		return nil, err
	}
	return rmF.FromResource(res), nil
}

// NewResMapFromSecretArgs takes a SecretArgs slice, generates
// secrets from each entry, and accumulates them in a ResMap.
func (rmF *Factory) NewResMapFromSecretArgs(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	argsList []types.SecretArgs) (ResMap, error) {
	var resources []*resource.Resource
	for _, args := range argsList {
		res, err := rmF.resF.MakeSecret(ldr, options, &args)
		if err != nil {
			return nil, errors.Wrap(err, "NewResMapFromSecretArgs")
		}
		resources = append(resources, res)
	}
	return newResMapFromResourceSlice(resources)
}

func (rmF *Factory) FromSecretArgs(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args types.SecretArgs) (ResMap, error) {
	res, err := rmF.resF.MakeSecret(ldr, options, &args)
	if err != nil {
		return nil, err
	}
	return rmF.FromResource(res), nil
}

func (rmF *Factory) MergePatches(patches []*resource.Resource) (
	ResMap, error) {
	return rmF.tf.MergePatches(patches, rmF.resF)
}

func newResMapFromResourceSlice(resources []*resource.Resource) (ResMap, error) {
	result := New()
	for _, res := range resources {
		err := result.Append(res)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
